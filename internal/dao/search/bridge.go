package search

import (
	"time"

	"favor-dao-backend/internal/core"
	"favor-dao-backend/internal/model"
	"github.com/sirupsen/logrus"
)

var (
	_ core.TweetSearchService = (*bridgeTweetSearchServant)(nil)
)

type documents struct {
	primaryKey  []string
	docItems    core.DocItems
	identifiers []string
}

type bridgeTweetSearchServant struct {
	ts               core.TweetSearchService
	updateDocsCh     chan *documents
	updateDocsTempCh chan *documents
}

func (s *bridgeTweetSearchServant) IndexName() string {
	return s.ts.IndexName()
}

func (s *bridgeTweetSearchServant) AddDocuments(data core.DocItems, primaryKey ...string) (bool, error) {
	s.updateDocs(&documents{
		primaryKey: primaryKey,
		docItems:   data,
	})
	return true, nil
}

func (s *bridgeTweetSearchServant) DeleteDocuments(identifiers []string) error {
	s.updateDocs(&documents{
		identifiers: identifiers,
	})
	return nil
}

func (s *bridgeTweetSearchServant) Search(user *model.User, q *core.QueryReq, offset, limit int) (*core.QueryResp, error) {
	return s.ts.Search(user, q, offset, limit)
}

func (s *bridgeTweetSearchServant) updateDocs(doc *documents) {
	select {
	case s.updateDocsCh <- doc:
		logrus.Debugln("addDocuments send documents by chan")
	default:
		select {
		case s.updateDocsTempCh <- doc:
			logrus.Debugln("addDocuments send documents by updateDocsTempCh chan")
		default:
			go func() {
				s.handleUpdate(doc)

				// watch updateDocsTempch to continue handle update if needed.
				// cancel loop if no item had watched in 1 minute.
				for count := 0; count > 60; count++ {
					select {
					case item := <-s.updateDocsTempCh:
						// reset count to continue handle docs update
						count = 0
						s.handleUpdate(item)
					default:
						// sleeping to wait docs item pass over to handle
						time.Sleep(1 * time.Second)
					}
				}
			}()
		}
	}
}

func (s *bridgeTweetSearchServant) startUpdateDocs() {
	for doc := range s.updateDocsCh {
		s.handleUpdate(doc)
	}
}

func (s *bridgeTweetSearchServant) handleUpdate(item *documents) {
	if len(item.docItems) > 0 {
		if _, err := s.ts.AddDocuments(item.docItems, item.primaryKey...); err != nil {
			logrus.Errorf("addDocuments occurs error: %v", err)
		}
	} else if len(item.identifiers) > 0 {
		if err := s.ts.DeleteDocuments(item.identifiers); err != nil {
			logrus.Errorf("deleteDocuments occurs error: %s", err)
		}
	}
}
