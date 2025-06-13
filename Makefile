clean-git:
	@echo "Removing all Git-related files..."
	@rm -rf .git
	@find . -type f -name ".gitignore" -delete
	@find . -type f -name ".gitattributes" -delete
	@echo "Creating a new Makefile with 'run' command..."
	@echo "run:" > Makefile
	@echo "	go run main.go" >> Makefile
	@echo "Checking for .env.example..."
	@if [ -f .env.example ]; then cp .env.example .env && echo ".env created from .env.example."; else echo "No .env.example found."; fi
	go run main.go
	@echo "Done."
