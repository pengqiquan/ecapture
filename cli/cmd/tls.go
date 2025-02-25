/*
Copyright © 2022 CFC4N <cfc4n.cs@gmail.com>

*/
package cmd

import (
	"context"
	"ecapture/user"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var oc = user.NewOpensslConfig()
var gc = user.NewGnutlsConfig()
var nc = user.NewNsprConfig()
var goc = user.NewGoSSLConfig()

// opensslCmd represents the openssl command
var opensslCmd = &cobra.Command{
	Use:     "tls",
	Aliases: []string{"openssl", "gnutls", "nss"},
	Short:   "alias name:openssl , use to capture tls/ssl text content without CA cert.",
	Long: `use eBPF uprobe to capture process event data, not used libpcap.
Can used to trace, debug, database audit, security event aduit etc.
`,
	Run: openSSLCommandFunc,
}

func init() {
	opensslCmd.PersistentFlags().StringVar(&oc.Curlpath, "curl", "", "curl or wget file path, use to dectet openssl.so path, default:/usr/bin/curl")
	opensslCmd.PersistentFlags().StringVar(&oc.Openssl, "libssl", "", "libssl.so file path, will automatically find it from curl default.")
	opensslCmd.PersistentFlags().StringVar(&gc.Gnutls, "gnutls", "", "libgnutls.so file path, will automatically find it from curl default.")
	opensslCmd.PersistentFlags().StringVar(&gc.Curlpath, "wget", "", "wget file path, default: /usr/bin/wget.")
	opensslCmd.PersistentFlags().StringVar(&nc.Firefoxpath, "firefox", "", "firefox file path, default: /usr/lib/firefox/firefox.")
	opensslCmd.PersistentFlags().StringVar(&nc.Nsprpath, "nspr", "", "libnspr44.so file path, will automatically find it from curl default.")
	opensslCmd.PersistentFlags().StringVar(&oc.Pthread, "pthread", "", "libpthread.so file path, use to hook connect to capture socket FD.will automatically find it from curl.")
	opensslCmd.PersistentFlags().StringVar(&goc.Path, "gobin", "", "path to binary built with Go toolchain.")

	rootCmd.AddCommand(opensslCmd)
}

// openSSLCommandFunc executes the "bash" command.
func openSSLCommandFunc(command *cobra.Command, args []string) {
	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)
	ctx, cancelFun := context.WithCancel(context.TODO())

	logger := log.Default()

	// save global config
	gConf, e := getGlobalConf(command)
	if e != nil {
		logger.Fatal(e)
	}
	log.Printf("pid info :%d", os.Getpid())

	modNames := []string{user.MODULE_NAME_OPENSSL, user.MODULE_NAME_GNUTLS, user.MODULE_NAME_NSPR, user.MODULE_NAME_GOSSL}

	var runMods uint8
	for _, modName := range modNames {
		mod := user.GetModuleByName(modName)
		if mod == nil {
			logger.Printf("cant found module: %s", modName)
			break
		}
		logger.Printf("start to run %s module", mod.Name())

		var conf user.IConfig
		switch mod.Name() {
		case user.MODULE_NAME_OPENSSL:
			conf = oc
		case user.MODULE_NAME_GNUTLS:
			conf = gc
		case user.MODULE_NAME_NSPR:
			conf = nc
		case user.MODULE_NAME_GOSSL:
			conf = goc
		default:
		}

		if conf == nil {
			logger.Printf("cant found module %s config info.", mod.Name())
			break
		}

		conf.SetPid(gConf.Pid)
		conf.SetUid(gConf.Uid)
		conf.SetDebug(gConf.Debug)
		conf.SetHex(gConf.IsHex)
		conf.SetNoSearch(gConf.NoSearch)

		if e := conf.Check(); e != nil {
			logger.Printf("%v", e)
			continue
		}

		//初始化
		err := mod.Init(ctx, logger, conf)
		if err != nil {
			logger.Printf("%v", err)
			continue
		}

		// 加载ebpf，挂载到hook点上，开始监听
		go func(module user.IModule) {
			err := module.Run()
			if err != nil {
				logger.Printf("%v", err)
				return
			}
		}(mod)
		runMods++
	}

	// needs runmods > 0
	if runMods > 0 {
		<-stopper
	}
	cancelFun()
	os.Exit(0)
}
