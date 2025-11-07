package summaryofversion

// GetDotfileByName returns a pointer to the Dotfile with the given name.
func GetDotfileByName(name string) *Dotfile {
	for _, d := range Registry {
		if d.Name == name {
			return &d
		}
	}
	return nil
}
