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

## Note
~~This program is heaviliy unoptimized. Every drop is running in its own goroutine. So...beware if you are on weaker hardware. Might optimize later. Probably could get away with running multiple drops per goroutine at the same speed without being noticeable.~~

Making drops not speed up to infinity, limiting goroutines to 50, and adding a 50ms delay in the render loop dramatically improved performance!
