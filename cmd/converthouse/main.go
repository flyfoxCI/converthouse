package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/deepfabric/converthouse/pkg/server"
	"github.com/deepfabric/converthouse/pkg/util"
	"github.com/fagongzi/log"
)

var (
	name                = flag.String("name", "node1", "Node name.")
	data                = flag.String("data", "/tmp/converthouse", "Converthouse data path")
	addr                = flag.String("addr", "127.0.0.1:8080", "Addr: http request entrypoint")
	addrReplica         = flag.String("addr-replica", "", "Addr: internal replica rpc addr")
	addrPPROF           = flag.String("addr-pprof", "", "Addr: pprof addr")
	limitCpus           = flag.Int("limit-cpus", 0, "Limit: schedule threads count")
	maxPeerDownTime     = flag.Int("max-peer-down", 60*5, "Seconds: allow a replication peer of the maximum offline time")
	resourceHBInterval  = flag.Int("interval-resource-hb", 5, "Seconds: resource heartbeat interval")
	containerHBInterval = flag.Int("interval-container-hb", 30, "Seconds: container heartbeat interval")
	resourceWorkerCount = flag.Uint64("worker-resources", 32, "Count: the number of worker for resource processing event")
	version             = flag.Bool("version", false, "Show version info")

	wait = flag.Int("wait", 0, "wait seconds before start")
)

func main() {
	flag.Parse()

	if *version {
		util.PrintVersion()
		os.Exit(0)
	}

	if *wait > 0 {
		time.Sleep(time.Second * time.Duration(*wait))
	}

	log.InitLog()
	util.InitProphetLog()

	if *limitCpus == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(*limitCpus)
	}

	if *addrPPROF != "" {
		go func() {
			log.Errorf("start pprof failed, errors:\n%+v",
				http.ListenAndServe(*addrPPROF, nil))
		}()
	}

	var cfg server.Cfg
	cfg.Store.Name = *name
	cfg.Store.DataPath = *data
	cfg.Store.AddrReplica = *addrReplica
	cfg.AddrHTTP = *addr
	cfg.Store.MaxPeerDownTime = *maxPeerDownTime
	cfg.Store.ResourceHeartbeatInterval = *resourceHBInterval
	cfg.Store.ContainerHeartbeatInterval = *containerHBInterval
	cfg.Store.ResourceWorkerCount = *resourceWorkerCount

	s := server.NewServer(cfg)
	go s.Start()
	waitStop(s)
}

func waitStop(s *server.Server) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sc
	log.Infof("exit: signal=<%d>.", sig)
	s.Stop()
	switch sig {
	case syscall.SIGTERM:
		log.Infof("exit: bye :-).")
		os.Exit(0)
	default:
		log.Infof("exit: bye :-(.")
		os.Exit(1)
	}
}
