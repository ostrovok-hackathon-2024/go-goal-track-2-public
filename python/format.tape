Output format.gif

Set FontSize 22
Set Width 1200
Set Height 800
Set Shell zsh
Set Theme "Catppuccin Frappe"

Type "poetry run cli -i ../inputs/rates_short.csv --categories bedding --categories quality -o out.json -f json && /bin/cat out.json | jq . | head"

Sleep 200ms

Enter

Sleep 8s
