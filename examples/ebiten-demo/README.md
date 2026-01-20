# MicroUI Go Demo - DEBUG VERSION

This is a **minimal debug test** to verify basic rendering works.

## Running

```bash
cd examples/ebiten-demo
go run main.go
```

## What You SHOULD See:

1. **Dark screen** (800x600)
2. **White text** at top-left showing:
   - Mouse position
   - Click count
   - Instructions
3. **BLUE RECTANGLE** at position (100, 100) - this is the button!
   - Size: 200x60 pixels
   - Color: Blue (RGB 60,120,200)
   - White text "TEST BUTTON" inside

## Testing:

1. **Move mouse** to position (100-300, 100-160)
2. **Click** inside the blue rectangle
3. **Watch console** - should log:
   ```
   MOUSE CLICK at 150,130
   BUTTON CLICKED! Count: 1
   ```
4. **Check screen** - Count should increment in the text at top

## What This Proves:

- ✅ If you see the **blue rectangle** at (100,100): **Rendering works!**
- ✅ If clicking increments counter: **Input handling works!**
- ❌ If you DON'T see blue rectangle: **Renderer is broken**
- ❌ If clicking doesn't work: **Mouse input is broken**

## Current Issues Being Debugged:

The `Button()` method and layout system have positioning issues.
This demo bypasses them to test the core rendering pipeline directly.

## Console Output:

When working correctly:
```
Mouse: 150,130
MOUSE CLICK at 150,130
BUTTON CLICKED! Count: 1
```

If you DON'T see the blue rectangle at (100,100) or clicking doesn't increment the counter, please report:
- What you see on screen
- Console output when you click
- Your OS and version
