package client

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	mlogger "g_matching_engine/pkg/mlog"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	prome "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const maxTxnOps = 128

type Op int

const (
	OpGet Op = iota
	OpDel
	OpGetPrefix
	OpDelPrefix
)

type Client struct {
	*clientv3.Client
	config *Config
}

func newClient(config *Config) *Client {
	conf := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.ConnectTimeout,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(prome.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(prome.StreamClientInterceptor),
		},
		//todo AutoSyncInterval
	}

	//config.logger = config.logger.With(xlog.FieldAddrAny(config.Endpoints))

	if config.Endpoints == nil {
		mlogger.Logger.Panic("client etcd endpoints empty :")
	}

	if !config.Secure {
		conf.DialOptions = append(conf.DialOptions, grpc.WithInsecure())
	}

	if config.BasicAuth {
		conf.Username = config.UserName
		conf.Password = config.Password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	//todo
	//CaCert
	//Certfile

	if tlsEnabled {
		conf.TLS = tlsConfig
	}

	client, err := clientv3.New(conf)

	if err != nil {
		mlogger.Logger.Panic("client etcd start panic:", zap.Any("", err))
	}

	cc := &Client{
		Client: client,
		config: config,
	}

	mlogger.Logger.Info("dial etcd server")
	return cc
}

func (client *Client) GetKeyValue(ctx context.Context, key string) (kv *mvccpb.KeyValue, err error) {
	rp, err := client.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(rp.Kvs) > 0 {
		return rp.Kvs[0], nil
	}

	return
}

//get prerfix
func (client *Client) GetPrefix(ctx context.Context, prefix string) (res map[string]string, err error) {

	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return
	}

	for _, kv := range resp.Kvs {
		res[string(kv.Key)] = string(kv.Value)
	}

	return
}

//delPrefix
func (client *Client) DelPrefix(ctx context.Context, prefix string) (int64, error) {
	resp, err := client.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}
	return resp.Deleted, err
}

func (client *Client) GetValues(ctx context.Context, keys ...string) (res map[string]string, err error) {
	var firstRevision int64 = 0
	var getOps = make([]string, 0, maxTxnOps)
	for i, key := range keys {
		getOps = append(getOps, key)
		//batch dotxn tx
		if len(getOps) >= maxTxnOps || i == len(keys)-1 {
			result, err := client.dotxn(OpGet, getOps, firstRevision, ctx)
			if err != nil {
				return nil, err
			}
			for i, r := range result.Responses {
				originKey := getOps[i]
				originKeyFixed := originKey
				if !strings.HasSuffix(originKeyFixed, "/") {
					originKeyFixed = originKey + "/"
				}
				for _, ev := range r.GetResponseRange().Kvs {
					k := string(ev.Key)
					if k == originKey || strings.HasPrefix(k, originKeyFixed) {
						res[string(ev.Key)] = string(ev.Value)
					}
				}
				if firstRevision == 0 {
					firstRevision = result.Header.GetRevision()
				}
			}
			getOps = getOps[:0]
		}
	}
	return
}
func dumpPrefixTx() {}
func (client *Client) dotxn(op Op, ops []string, firstRevision int64, ctx context.Context) (result *clientv3.TxnResponse, err error) {
	txnOps := make([]clientv3.Op, 0, maxTxnOps)

	for _, k := range ops {
		switch op {
		case OpGet:
			txnOps = append(
				txnOps, clientv3.OpGet(k,
					//clientv3.WithPrefix(),
					clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
					clientv3.WithRev(firstRevision)))
		case OpGetPrefix:
			txnOps = append(
				txnOps, clientv3.OpGet(k,
					clientv3.WithPrefix(),
					clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
					clientv3.WithRev(firstRevision)))
		case OpDel:
			txnOps = append(
				txnOps, clientv3.OpDelete(k,
					//clientv3.WithPrefix(),
					clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
					clientv3.WithRev(firstRevision)))
		case OpDelPrefix:
			txnOps = append(
				txnOps, clientv3.OpDelete(k,
					clientv3.WithPrefix(),
					clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
					clientv3.WithRev(firstRevision)))
		}
	}
	result, err = client.Txn(ctx).Then(txnOps...).Commit()
	if err != nil {
		return result, err
	}
	return result, err

}
