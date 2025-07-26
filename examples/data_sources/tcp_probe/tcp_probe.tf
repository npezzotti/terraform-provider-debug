data "debug_tcp_probe" "example" {
  host    = "example.com"
  port    = 80
  timeout = 2
}
