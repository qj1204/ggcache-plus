package global

import (
	"ggcache-plus/config"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

var (
	Config            *config.Config
	DB                *gorm.DB
	Log               *logrus.Logger
	DefaultEtcdConfig clientv3.Config
)
