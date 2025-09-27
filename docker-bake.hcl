variable "REGISTRY" {
  default = "ghcr.io"
}

variable "REPOSITORY" {
  default = "elek/smokeping"
}

variable "TAG" {
  default = "latest"
}

group "default" {
  targets = ["build"]
}

target "build" {
  dockerfile = "Dockerfile"
  context = "."
  tags = [
    "${REGISTRY}/${REPOSITORY}:${TAG}",
    "${REGISTRY}/${REPOSITORY}:latest"
  ]
  platforms = ["linux/amd64"]
}

target "smokeping-local" {
  inherits = ["smokeping"]
  tags = ["smokeping:local"]
  platforms = ["linux/amd64"]
}

target "smokeping-push" {
  inherits = ["smokeping"]
  output = ["type=registry"]
}