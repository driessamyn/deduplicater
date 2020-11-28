package main

import "github.com/manifoldco/promptui"

func PromptAction(dirValidateFunc func(target string) error) (FindAction, *string) {
	const CANCEL = "Do nothing"
	const DELETE = "Delete duplicates"
	const MOVE = "Move files to another folder"
	items := []string{CANCEL, DELETE, MOVE}

	prompt := promptui.Select{
		Label: "What do you want to do with the duplucates?",
		Items: items,
	}

	_, result, _ := prompt.Run()

	switch result {
	case DELETE:
		confirmPrompt := promptui.Prompt{
			Label:     "Are you sure you want to permanantly delete duplicates? (THIS CANNOT BE UNDONE)",
			IsConfirm: true,
		}

		if confirm, _ := confirmPrompt.Run(); "Y" == confirm {
			return Delete, nil
		}
	case MOVE:
		movePrompt := promptui.Prompt{
			Label:    "Location to copy duplicate files to",
			Validate: dirValidateFunc,
		}

		moveDir, _ := movePrompt.Run()
		return Move, &moveDir
	}
	return Unknown, nil
}
