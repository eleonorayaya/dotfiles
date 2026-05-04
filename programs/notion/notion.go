package notion

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "notion"
}
