package theme

func preset(name string) Theme {
	switch name {
	case "catppuccin-mocha":
		return catppuccinMocha()
	case "catppuccin-latte":
		return catppuccinLatte()
	case "dracula":
		return dracula()
	case "nord":
		return nord()
	case "gruvbox-dark":
		return gruvboxDark()
	case "gruvbox-light":
		return gruvboxLight()
	case "one-dark":
		return oneDark()
	case "solarized-dark":
		return solarizedDark()
	case "solarized-light":
		return solarizedLight()
	case "light":
		return lightTheme()
	default: // "dark" and fallback
		return darkTheme()
	}
}

func darkTheme() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#569cd6", String: "#ce9178", Comment: "#6a9955",
			Function: "#dcdcaa", FunctionMacro: "#c586c0", Type: "#4ec9b0",
			Number: "#b5cea8", Operator: "#d4d4d4", Variable: "#9cdcfe",
			VariableBuiltin: "#569cd6", VariableMember: "#9cdcfe",
			Module: "#4ec9b0", Tag: "#569cd6", Attribute: "#9cdcfe",
			Label: "#c586c0", Punctuation: "#d4d4d4",
		},
		Diff: DiffColors{
			AddedBg: "#1a2e1a", DeletedBg: "#2e1a1a",
			AddedWordBg: "#2d4a2d", DeletedWordBg: "#4a2d2d",
			AddedGutterBg: "#163016", DeletedGutterBg: "#301616",
		},
		UI: UIColors{
			Border: "#444444", Text: "#d4d4d4", Selection: "#264f78",
			SearchHighlight: "#515c6a", StatusAdded: "#6a9955",
			StatusModified: "#dcdcaa", StatusDeleted: "#f44747",
		},
	}
}

func lightTheme() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#0000ff", String: "#a31515", Comment: "#008000",
			Function: "#795e26", FunctionMacro: "#af00db", Type: "#267f99",
			Number: "#098658", Operator: "#000000", Variable: "#001080",
			VariableBuiltin: "#0000ff", VariableMember: "#001080",
			Module: "#267f99", Tag: "#800000", Attribute: "#ff0000",
			Label: "#af00db", Punctuation: "#000000",
		},
		Diff: DiffColors{
			AddedBg: "#dafada", DeletedBg: "#fadada",
			AddedWordBg: "#b5f0b5", DeletedWordBg: "#f0b5b5",
			AddedGutterBg: "#c8f0c8", DeletedGutterBg: "#f0c8c8",
		},
		UI: UIColors{
			Border: "#cccccc", Text: "#000000", Selection: "#add6ff",
			SearchHighlight: "#d6ebff", StatusAdded: "#008000",
			StatusModified: "#795e26", StatusDeleted: "#cd3131",
		},
	}
}

func catppuccinMocha() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#cba6f7", String: "#a6e3a1", Comment: "#585b70",
			Function: "#89b4fa", FunctionMacro: "#cba6f7", Type: "#f38ba8",
			Number: "#fab387", Operator: "#cdd6f4", Variable: "#cdd6f4",
			VariableBuiltin: "#f38ba8", VariableMember: "#cdd6f4",
			Module: "#89b4fa", Tag: "#f38ba8", Attribute: "#fab387",
			Label: "#f38ba8", Punctuation: "#cdd6f4",
		},
		Diff: DiffColors{
			AddedBg: "#1e3a2a", DeletedBg: "#3a1e1e",
			AddedWordBg: "#2d5a3d", DeletedWordBg: "#5a2d2d",
			AddedGutterBg: "#182e20", DeletedGutterBg: "#2e1818",
		},
		UI: UIColors{
			Border: "#313244", Text: "#cdd6f4", Selection: "#45475a",
			SearchHighlight: "#494d64", StatusAdded: "#a6e3a1",
			StatusModified: "#f9e2af", StatusDeleted: "#f38ba8",
		},
	}
}

func catppuccinLatte() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#8839ef", String: "#40a02b", Comment: "#9ca0b0",
			Function: "#1e66f5", FunctionMacro: "#8839ef", Type: "#d20f39",
			Number: "#fe640b", Operator: "#4c4f69", Variable: "#4c4f69",
			VariableBuiltin: "#d20f39", VariableMember: "#4c4f69",
			Module: "#1e66f5", Tag: "#d20f39", Attribute: "#fe640b",
			Label: "#8839ef", Punctuation: "#4c4f69",
		},
		Diff: DiffColors{
			AddedBg: "#d8f0de", DeletedBg: "#f0d8d8",
			AddedWordBg: "#b8e8c0", DeletedWordBg: "#e8b8b8",
			AddedGutterBg: "#c8e8d0", DeletedGutterBg: "#e8c8c8",
		},
		UI: UIColors{
			Border: "#ccd0da", Text: "#4c4f69", Selection: "#bcc0cc",
			SearchHighlight: "#dce0e8", StatusAdded: "#40a02b",
			StatusModified: "#df8e1d", StatusDeleted: "#d20f39",
		},
	}
}

func dracula() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#ff79c6", String: "#f1fa8c", Comment: "#6272a4",
			Function: "#50fa7b", FunctionMacro: "#ff79c6", Type: "#8be9fd",
			Number: "#bd93f9", Operator: "#f8f8f2", Variable: "#f8f8f2",
			VariableBuiltin: "#ff79c6", VariableMember: "#f8f8f2",
			Module: "#8be9fd", Tag: "#ff79c6", Attribute: "#50fa7b",
			Label: "#ff79c6", Punctuation: "#f8f8f2",
		},
		Diff: DiffColors{
			AddedBg: "#1a3020", DeletedBg: "#30201a",
			AddedWordBg: "#2a4830", DeletedWordBg: "#48302a",
			AddedGutterBg: "#162818", DeletedGutterBg: "#281816",
		},
		UI: UIColors{
			Border: "#44475a", Text: "#f8f8f2", Selection: "#44475a",
			SearchHighlight: "#6272a4", StatusAdded: "#50fa7b",
			StatusModified: "#f1fa8c", StatusDeleted: "#ff5555",
		},
	}
}

func nord() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#81a1c1", String: "#a3be8c", Comment: "#616e88",
			Function: "#88c0d0", FunctionMacro: "#b48ead", Type: "#8fbcbb",
			Number: "#b48ead", Operator: "#eceff4", Variable: "#d8dee9",
			VariableBuiltin: "#81a1c1", VariableMember: "#d8dee9",
			Module: "#8fbcbb", Tag: "#81a1c1", Attribute: "#88c0d0",
			Label: "#b48ead", Punctuation: "#eceff4",
		},
		Diff: DiffColors{
			AddedBg: "#1e2e1e", DeletedBg: "#2e1e1e",
			AddedWordBg: "#2d422d", DeletedWordBg: "#42302d",
			AddedGutterBg: "#192619", DeletedGutterBg: "#261919",
		},
		UI: UIColors{
			Border: "#3b4252", Text: "#eceff4", Selection: "#434c5e",
			SearchHighlight: "#4c566a", StatusAdded: "#a3be8c",
			StatusModified: "#ebcb8b", StatusDeleted: "#bf616a",
		},
	}
}

func gruvboxDark() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#fb4934", String: "#b8bb26", Comment: "#928374",
			Function: "#fabd2f", FunctionMacro: "#fe8019", Type: "#8ec07c",
			Number: "#d3869b", Operator: "#ebdbb2", Variable: "#ebdbb2",
			VariableBuiltin: "#fb4934", VariableMember: "#ebdbb2",
			Module: "#83a598", Tag: "#fb4934", Attribute: "#fabd2f",
			Label: "#fe8019", Punctuation: "#ebdbb2",
		},
		Diff: DiffColors{
			AddedBg: "#1d2b1d", DeletedBg: "#2b1d1d",
			AddedWordBg: "#2d402d", DeletedWordBg: "#40302d",
			AddedGutterBg: "#192319", DeletedGutterBg: "#231919",
		},
		UI: UIColors{
			Border: "#504945", Text: "#ebdbb2", Selection: "#3c3836",
			SearchHighlight: "#504945", StatusAdded: "#b8bb26",
			StatusModified: "#fabd2f", StatusDeleted: "#fb4934",
		},
	}
}

func gruvboxLight() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#9d0006", String: "#79740e", Comment: "#928374",
			Function: "#b57614", FunctionMacro: "#af3a03", Type: "#427b58",
			Number: "#8f3f71", Operator: "#3c3836", Variable: "#3c3836",
			VariableBuiltin: "#9d0006", VariableMember: "#3c3836",
			Module: "#076678", Tag: "#9d0006", Attribute: "#b57614",
			Label: "#af3a03", Punctuation: "#3c3836",
		},
		Diff: DiffColors{
			AddedBg: "#daeada", DeletedBg: "#eadada",
			AddedWordBg: "#b8d8b8", DeletedWordBg: "#d8b8b8",
			AddedGutterBg: "#c8d8c8", DeletedGutterBg: "#d8c8c8",
		},
		UI: UIColors{
			Border: "#d5c4a1", Text: "#3c3836", Selection: "#ebdbb2",
			SearchHighlight: "#f2e5bc", StatusAdded: "#79740e",
			StatusModified: "#b57614", StatusDeleted: "#9d0006",
		},
	}
}

func oneDark() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#c678dd", String: "#98c379", Comment: "#5c6370",
			Function: "#61afef", FunctionMacro: "#c678dd", Type: "#e5c07b",
			Number: "#d19a66", Operator: "#abb2bf", Variable: "#abb2bf",
			VariableBuiltin: "#e06c75", VariableMember: "#abb2bf",
			Module: "#61afef", Tag: "#e06c75", Attribute: "#d19a66",
			Label: "#c678dd", Punctuation: "#abb2bf",
		},
		Diff: DiffColors{
			AddedBg: "#1d2b1d", DeletedBg: "#2b1d1d",
			AddedWordBg: "#2d3f2d", DeletedWordBg: "#3f2d2d",
			AddedGutterBg: "#192519", DeletedGutterBg: "#251919",
		},
		UI: UIColors{
			Border: "#3e4452", Text: "#abb2bf", Selection: "#3e4452",
			SearchHighlight: "#3e4452", StatusAdded: "#98c379",
			StatusModified: "#e5c07b", StatusDeleted: "#e06c75",
		},
	}
}

func solarizedDark() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#859900", String: "#2aa198", Comment: "#586e75",
			Function: "#268bd2", FunctionMacro: "#d33682", Type: "#b58900",
			Number: "#d33682", Operator: "#839496", Variable: "#839496",
			VariableBuiltin: "#859900", VariableMember: "#839496",
			Module: "#268bd2", Tag: "#859900", Attribute: "#cb4b16",
			Label: "#d33682", Punctuation: "#839496",
		},
		Diff: DiffColors{
			AddedBg: "#1a2b1a", DeletedBg: "#2b1a1a",
			AddedWordBg: "#253c25", DeletedWordBg: "#3c2525",
			AddedGutterBg: "#162416", DeletedGutterBg: "#241616",
		},
		UI: UIColors{
			Border: "#073642", Text: "#839496", Selection: "#073642",
			SearchHighlight: "#073642", StatusAdded: "#859900",
			StatusModified: "#b58900", StatusDeleted: "#dc322f",
		},
	}
}

func solarizedLight() Theme {
	return Theme{
		Syntax: SyntaxColors{
			Keyword: "#859900", String: "#2aa198", Comment: "#93a1a1",
			Function: "#268bd2", FunctionMacro: "#d33682", Type: "#b58900",
			Number: "#d33682", Operator: "#657b83", Variable: "#657b83",
			VariableBuiltin: "#859900", VariableMember: "#657b83",
			Module: "#268bd2", Tag: "#859900", Attribute: "#cb4b16",
			Label: "#d33682", Punctuation: "#657b83",
		},
		Diff: DiffColors{
			AddedBg: "#daeada", DeletedBg: "#eadada",
			AddedWordBg: "#b8d8b8", DeletedWordBg: "#d8b8b8",
			AddedGutterBg: "#c8d8c8", DeletedGutterBg: "#d8c8c8",
		},
		UI: UIColors{
			Border: "#eee8d5", Text: "#657b83", Selection: "#eee8d5",
			SearchHighlight: "#fdf6e3", StatusAdded: "#859900",
			StatusModified: "#b58900", StatusDeleted: "#dc322f",
		},
	}
}
