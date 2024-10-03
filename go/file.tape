Output file.gif

Set FontSize 22
Set Width 1200
Set Height 800
Set Shell zsh
Set Theme "Catppuccin Frappe"

Type "go run cmd/cli/main.go -i ../inputs/rates_short.csv -c bedding -c quality -o out.csv && echo out.csv | head out.csv | xsv table"

Sleep 200ms

Enter

Sleep 8s
