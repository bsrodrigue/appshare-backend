package memory

// Container holds all in-memory repositories.
// This is useful for testing services that depend on multiple repositories.
type Container struct {
	User    *UserRepository
	Project *ProjectRepository
}

// NewContainer creates a new container with all repositories initialized.
func NewContainer() *Container {
	return &Container{
		User:    NewUserRepository(),
		Project: NewProjectRepository(),
	}
}

// Reset clears all data in all repositories.
func (c *Container) Reset() {
	c.User.Reset()
	c.Project.Reset()
}
