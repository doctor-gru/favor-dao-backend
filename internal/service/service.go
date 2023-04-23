package service

import (
	"fmt"
	"strings"

	"favor-dao-backend/internal/conf"
	"favor-dao-backend/internal/core"
	"favor-dao-backend/internal/dao"
	"favor-dao-backend/internal/model"
	"favor-dao-backend/pkg/comet"
	"favor-dao-backend/pkg/firebase"
	"favor-dao-backend/pkg/pointSystem"
	"favor-dao-backend/pkg/psub"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ds             core.DataService
	ts             core.TweetSearchService
	eth            *ethclient.Client
	chat           *comet.ChatGateway
	point          *pointSystem.Gateway
	pubsub         *psub.Service
	notifyFirebase *firebase.Client
)

func Initialize() {
	ds = dao.DataService()
	ts = dao.TweetSearchService()

	pubsub = psub.New()
	// MUST connect!
	client, err := ethclient.Dial(conf.EthSetting.Endpoint)
	if err != nil {
		panic(fmt.Sprintf("dial eth: %s", err))
	}
	eth = client
	notifyFirebase, err = firebase.New(conf.FirebaseSetting.Config)
	if err != nil {
		panic(err)
	}
	chat = comet.New(conf.ChatSetting.AppId, conf.ChatSetting.Region, conf.ChatSetting.ApiKey)
	point = pointSystem.New(conf.PointSetting.Gateway)
	conf.PointSetting.Callback = strings.TrimRight(conf.PointSetting.Callback, "/")
}

func persistMediaContents(contents []*PostContentItem) (items []string, err error) {
	items = make([]string, 0, len(contents))
	for _, item := range contents {
		switch item.Type {
		case model.CONTENT_TYPE_IMAGE,
			model.CONTENT_TYPE_VIDEO,
			model.CONTENT_TYPE_AUDIO:
			items = append(items, item.Content)
			if err != nil {
				continue
			}
		}
	}
	return
}
