terraform {
  required_version = ">= 1.6.0"
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4.0"
    }
  }
}

provider "docker" {}

variable "container_name" {
  type    = string
  default = "tofu-demo-nginx"
}

variable "host_port" {
  type    = number
  default = 8085
}

resource "docker_image" "nginx" {
  name         = "nginx:1.25-alpine"
  keep_locally = false
}

resource "docker_container" "nginx_app" {
  image = docker_image.nginx.image_id
  name  = var.container_name
  ports {
    internal = 80
    external = var.host_port
  }
}

resource "local_file" "ansible_inventory" {
  content = templatefile("${path.module}/inventory.tftpl", {
    server_ip       = "localhost"
    connection_type = "local"
    environment     = "production-lab"
  })
  filename        = "${path.module}/../ansible/hosts_generated.ini"
  file_permission = "0644"
}

output "app_url" {
  value = "http://localhost:${var.host_port}"
}

output "generated_inventory_path" {
  value = resource.local_file.ansible_inventory.filename
}
