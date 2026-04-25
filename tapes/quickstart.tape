Output tapes/quickstart.gif

Require himo
Require bat
Require starship

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 80ms
Set WindowBar "Colorful"
Set WindowBarSize 40

Env XDG_CONFIG_HOME "/tmp/himo-demo/config"
Env XDG_STATE_HOME "/tmp/himo-demo/state"
Env HIMO_DIR "/tmp/himo-demo/todos"
Env STARSHIP_CONFIG "/tmp/himo-demo/starship.toml"

Hide
Type "rm -rf /tmp/himo-demo/config /tmp/himo-demo/state /tmp/himo-demo/todos && mkdir -p /tmp/himo-demo/config/himo /tmp/himo-demo/todos && cd /tmp/himo-demo"
Enter
Type `printf 'base_dir = "/tmp/himo-demo/todos"\ndefault_project = "work"\npreview_pane = true\n' > /tmp/himo-demo/config/himo/config.toml`
Enter
Type `printf 'add_newline = false\nformat = "$character"\n[character]\nsuccess_symbol = "[\xe2\x9d\xaf](bold purple)"\nerror_symbol = "[\xe2\x9d\xaf](bold red)"\n' > /tmp/himo-demo/starship.toml`
Enter
Type `eval "$(starship init bash)"`
Enter
Type `bat() { command bat --paging=never --style=plain --theme="Catppuccin Mocha" --language=md "$@"; }`
Enter
Type "clear"
Enter
Show

Type "himo new work"
Enter
Sleep 600ms

Type "himo new home"
Enter
Sleep 800ms

Type "himo work"
Enter
Sleep 1500ms

Type "?"
Sleep 4000ms
Type "?"
Sleep 800ms

Type "o"
Sleep 400ms
Type "Write design doc"
Enter
Sleep 700ms

Type "o"
Sleep 400ms
Type "Migrate the payments table"
Enter
Sleep 700ms

Type "o"
Sleep 400ms
Type "Buy groceries"
Enter
Sleep 1000ms

Type "0"
Sleep 1000ms

Space
Sleep 900ms

Type "j"
Sleep 400ms
Type "!"
Sleep 900ms

Type "j"
Sleep 400ms
Type "x"
Sleep 1100ms

Type "2"
Sleep 1200ms
Type "3"
Sleep 1200ms
Type "4"
Sleep 1200ms
Type "5"
Sleep 1200ms
Type "0"
Sleep 800ms

Tab
Sleep 1200ms

Type "o"
Sleep 400ms
Type "Plan weekend trip"
Enter
Sleep 1000ms

Type "A"
Sleep 2000ms

Type "q"
Sleep 600ms

Hide
Type "cd /tmp/himo-demo/todos/work"
Enter
Type "clear"
Enter
Show

Type "bat active.md"
Enter
Sleep 2500ms

Type "bat done.md"
Enter
Sleep 3000ms
