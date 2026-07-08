variable "db_url" {
  type    = string
  default = "mysql://root:root@grit-mysql:3306/grit"
}

env "local" {
  url     = var.db_url
  dev     = "mysql://root:root@grit-mysql:3306/dev"
  src     = "file://sql/"
  schemas = ["grit"]
}

