# Utopia

Simple, git based and language agnostic template generator

![Demo](.github/demo.gif)

## Install
```shell script
go get -u github.com/gabrielcolson/utopia
```

## Usage
You can create a new project from an utopia github repository:
```shell script
utopia mySuperProject https://github.com/gabrielcolson/utopia-example.git 
```


To create a template, just add a `.utopia.yml` file at the root
of your repository:
```yaml
features:
  - name: eslint
    description: lint your project with eslint
    branch: utopia/eslint
  - name: prettier
    description: format your project with prettier
    branch: utopia/prettier
  - name: auth
    description: basic cookie based authentication
    branch: utopia/auth
```

For more details, take a look at the [example template](https://github.com/gabrielcolson/utopia-example)