package gitsyncer

type RemoteRepo struct {
	URL      string
	Username string
	Password string
}

type Task struct {
	Src RemoteRepo
	Dst RemoteRepo
}

type Config struct {
	Tasks []Task
}
