Output demo.gif

Set Margin 20
Set MarginFill "#009B77"
Set BorderRadius 10

Set FontSize 24
Set Width 1200
Set Height 600

Type "echo "
Sleep 100ms
Hide
Type "r8_•••••••••••••••••••••••••••••••••••••"
Show
Sleep 100ms
Type " | replicate auth login"
Sleep 100ms
Ctrl+C # Don't actually set the API key
Sleep 1s

Type 'replicate run meta/llama-2-70b-chat \'
Enter
Type@50ms '  prompt="write a haiku about corgis"'
Enter

Sleep 2s

Enter

Enter
Type 'replicate run stability-ai/sdxl \'
Enter
Type@50ms '  prompt="a studio photo of a rainbow colored corgi" \'
Enter
Type@50ms '  width=512 height=512 seed=42069'
Enter

Sleep 30s
