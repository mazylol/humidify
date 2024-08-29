# Humidify
Rain in the terminal

![](https://github.com/mazylol/humidify/blob/master/demo.gif)

## Install
### AUR
```bash
yay -S humidify-git
```

### Source
```bash
go install github.com/mazylol/humidify
```

## Run
```bash
humidify
```

## Features
- [x] Change raindrop character (--character)
- [x] Change raindrop speed (--speed)
- [x] Change raindrop color (--color)
- [x] Control number of raindrops falling (--density)
- [ ] Change splash size (--splash-size)
- [x] Set width of the terminal grid (--width)
- [x] Set height of the terminal grid (--height)
- [x] Change the gravity, or acceleration of the drops (--gravity)
- [x] Disable splash (--no-splash)

## Note
~~This program is heaviliy unoptimized. Every drop is running in its own goroutine. So...beware if you are on weaker hardware. Might optimize later. Probably could get away with running multiple drops per goroutine at the same speed without being noticeable.~~

Making drops not speed up to infinity, limiting goroutines to 50, and adding a 50ms delay in the render loop dramatically improved performance!
