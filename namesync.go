package main
import "github.com/hlandau/degoutils/config"
import "github.com/hlandau/degoutils/service"
import "github.com/hlandau/namesync/server"

func main() {
	cfg := server.Config{}
	config := config.Configurator{
		ProgramName: "namesync",
		ConfigFilePaths: []string{"$BIN/../etc/namesync.conf", "/etc/namesync/namesync.conf"},
	}
	config.ParseFatal(&cfg)

	service.Main(&service.Info{
		Name: "namesync",
		Description: "Namecoin to SQL Database Synchronization Daemon",
		DefaultChroot: service.EmptyChrootPath,
		RunFunc: func(smgr service.Manager) error {
			doneChan := make(chan error)

			cfg.StatusUpdateFunc = func(status string) {
				smgr.SetStatus("namesync: " + status)
			}

			cfg.StatusUpdateFunc("starting")

			go func() {
				err := server.Run(cfg, func() error {
					err := smgr.DropPrivileges()
					if err != nil {
						return err
					}

					smgr.SetStarted()

					return nil
				})
				doneChan <- err
			}()

			select {
				case <-smgr.StopChan():
					// Stop was requested. Just return, everything in the daemon is transactional
					// so we don't need to worry about a clean shutdown.
					return nil

				case err := <-doneChan:
					// Daemon stopped spontaneously. Run() never returns nil, though.
					if err == nil {
						panic("unreachable")
					}
					return err
			}

			return nil
		},
	})
}