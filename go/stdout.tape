Output stdout.gif

Set FontSize 22
Set Width 1200
Set Height 800
Set Shell zsh
Set Theme "Catppuccin Frappe"

Type "go run cmd/cli/main.go -i 'Standard - Double Room - With Breakfast' | xsv table"

Sleep 200ms

Enter

Sleep 8s

Hide
Type "clear"
Enter

Show

Type "go run cmd/cli/main.go -i ../inputs/rates_short.csv | xsv table"

Sleep 200ms

Enter

Sleep 8s
