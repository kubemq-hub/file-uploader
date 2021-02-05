package source

import (
	"context"
	"github.com/kubemq-io/file-uploader/config"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/kubemq-io/file-uploader/types"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Service struct {
	cfg        *config.Config
	waiting    sync.Map
	inProgress sync.Map
	completed  sync.Map
	senders    []types.Sender
	sendCh     chan *types.SourceFile
	logger     *logger.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewSourceService(cfg *config.Config) *Service {
	s := &Service{
		cfg:        cfg,
		waiting:    sync.Map{},
		inProgress: sync.Map{},
		completed:  sync.Map{},
		logger:     logger.NewLogger("source-service"),
	}
	return s
}
func (s *Service) Start(ctx context.Context, senders []types.Sender) {
	s.ctx, s.cancelFunc = context.WithCancel(ctx)
	s.senders = senders
	s.sendCh = make(chan *types.SourceFile)
	go s.scan(s.ctx)
	for i := 0; i < len(senders); i++ {
		go s.senderFunc(s.ctx, senders[i])
	}
	go s.send(s.ctx)
	s.logger.Info("starting source service")
}
func (s *Service) Stop() {
	s.cancelFunc()
	s.waiting = sync.Map{}
	s.logger.Info("source service stopped")
}
func (s *Service) inPipe(file *types.SourceFile) bool {
	if _, ok := s.waiting.Load(file.FullPath()); ok {
		return true
	}
	if _, ok := s.inProgress.Load(file.FullPath()); ok {
		return true
	}
	if _, ok := s.completed.Load(file.FullPath()); ok {
		s.logger.Infof("file %s already sent and will be deleted", file.FullPath())
		if err := file.Delete(); err != nil {
			s.logger.Errorf("error during delete a file %s,%s, will try again", file.FullPath(), err.Error())
		}
		return true
	}
	return false
}
func (s *Service) walk() error {
	var list []*types.SourceFile
	err := filepath.Walk(s.cfg.Source.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			list = append(list, types.NewSourceFile(info, path, s.cfg.Source.Root))
		}
		return nil
	})

	if err != nil {
		return err
	}
	added := 0
	for _, file := range list {
		if !s.inPipe(file) {
			s.waiting.Store(file.FullPath(), file)
			added++
		}
	}
	if added > 0 {
		s.logger.Infof("%d new files added to sending waiting list", added)
	}
	return nil
}

func (s *Service) senderFunc(ctx context.Context, sender types.Sender) {
	for {
		select {
		case file := <-s.sendCh:
			s.inProgress.Store(file.FullPath(), file)
			s.waiting.Delete(file.FullPath())
			req, err := file.Request(s.cfg.Source.BucketType, s.cfg.Source.BucketName)
			if err != nil {
				s.logger.Errorf("error during creating file requests %s, %s", file.FullPath(), err.Error())
				s.waiting.Store(file.FullPath(), file)
				s.inProgress.Delete(file.FullPath())
				continue
			}
			s.logger.Infof("sending file %s started", file.FileName())
			resp, err := sender.Send(ctx, req)
			if err != nil {
				s.logger.Errorf("error during sending file %s, %s", file.FileName(), err.Error())
				s.waiting.Store(file.FullPath(), file)
				s.inProgress.Delete(file.FullPath())
				continue
			}
			if resp.IsError {
				s.logger.Errorf("error on sending file %s response, %s", file.FileName(), resp.Error)
				s.waiting.Store(file.FullPath(), file)
				s.inProgress.Delete(file.FullPath())
				continue
			}
			if err := file.Delete(); err != nil {
				s.logger.Errorf("error during delete a file %s, %s,file will be resend", file.FileName(), err.Error())
				s.waiting.Store(file.FullPath(), file)
			} else {
				s.completed.Store(file.FullPath(), file)
			}
			s.inProgress.Delete(file.FullPath())
			s.logger.Infof("sending file %s completed", file.FileName())
		case <-ctx.Done():
			return
		}
	}
}
func (s *Service) scan(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Duration(s.cfg.Source.PollIntervalSeconds) * time.Second):
			err := s.walk()
			if err != nil {
				s.logger.Errorf("error during scan files, %s", err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}
func (s *Service) send(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Second):
			var list []*types.SourceFile
			s.waiting.Range(func(key, value interface{}) bool {
				list = append(list, value.(*types.SourceFile))
				return true
			})
			for _, file := range list {
				if _, ok := s.inProgress.Load(file.FullPath()); ok {
					continue
				}
				select {
				case s.sendCh <- file:
				case <-ctx.Done():
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
func (s *Service) Status() *Status {
	st := &Status{
		Waiting:    0,
		InProgress: 0,
		Completed:  0,
	}
	s.waiting.Range(func(key, value interface{}) bool {
		st.Waiting++
		return true
	})

	s.inProgress.Range(func(key, value interface{}) bool {
		st.InProgress++
		return true
	})
	s.completed.Range(func(key, value interface{}) bool {
		st.Completed++
		return true
	})
	return st
}
