# Humidify
Rain in the terminal

![](https://github.com/mazylol/humidify/blob/master/demo.gif)

## Install
```bash
go install github.com/mazylol/humidify
```

## Run
```bash
humidify
```

## Note
This program is heaviliy unoptimized. Every drop is running in its own goroutine. So...beware if you are on weaker hardware. Might optimize later. Probably could get away with running multiple drops per goroutine at the same speed without being noticeable.
