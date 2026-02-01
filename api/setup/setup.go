package setup

func InitProject(cfg Config) {
  initDatabase(cfg)
  initRouters(cfg)
}
