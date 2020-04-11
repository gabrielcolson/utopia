# Utopia

Simple, git based and language agnostic boilerplate generator

![Demo](.github/demo.gif)

## Install
```shell script
go get -u github.com/gabrielcolson/utopia
```

## Usage
You can create a new project from a utopia github repository:
```shell script
utopia https://github.com/gabrielcolson/utopia-example.git 
```


To create a template, just add a `.utopia.yml` file at the root
of your repository:
```yaml
features:
  eslint:
    branch: utopia/eslint
  prettier:
    branch: utopia/prettier
  auth:
    branch: utopia/auth
```

For more details, take a look at the [example template](https://github.com/gabrielcolson/utopia-example)