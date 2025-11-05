package dotfiles

type Dotfile struct {
	Name         string
	Author       string
	Repo         string
	Branch       string
	HasReleases  bool
	Description  string
}

var Registry = []Dotfile{
	{
		Name:        "HyDE",
		Author:      "HyDE-Project",
		Repo:        "https://github.com/HyDE-Project/HyDE",
		Branch:      "master",
		HasReleases: false,
		Description: "Dynamic modular Hyprland setup.",
	},
	{
		Name:        "Hyprdots",
		Author:      "prasanthrangan",
		Repo:        "https://github.com/prasanthrangan/hyprdots",
		Branch:      "main",
		HasReleases: false,
		Description: "Full-featured Arch-based Hyprland dotfiles.",
	},
	{
		Name:        "JaKooLit-Dots",
		Author:      "JaKooLit",
		Repo:        "https://github.com/JaKooLit/Hyprland-Dots",
		Branch:      "main",
		HasReleases: false,
		Description: "Multi-distro prebuilt Hyprland configurations.",
	},
	{
		Name:        "end4-dots",
		Author:      "end-4",
		Repo:        "https://github.com/end-4/dots-hyprland",
		Branch:      "main",
		HasReleases: false,
		Description: "User-centric, aesthetic rice.",
	},
	{
		Name:        "ML4W-Dotfiles",
		Author:      "mylinuxforwork",
		Repo:        "https://github.com/mylinuxforwork/dotfiles",
		Branch:      "main",
		HasReleases: true,
		Description: "Production-ready Hyprland workspace.",
	},
	{
		Name:        "taylor-hyprland",
		Author:      "taylor85345",
		Repo:        "https://github.com/taylor85345/hyprland-dotfiles",
		Branch:      "main",
		HasReleases: false,
		Description: "Minimal Hyprland config for Arch users.",
	},
	{
		Name:        "elifouts-dotfiles",
		Author:      "elifouts",
		Repo:        "https://github.com/elifouts/Dotfiles",
		Branch:      "main",
		HasReleases: false,
		Description: "Clean Hyprland environment with GTK theming.",
	},
}
