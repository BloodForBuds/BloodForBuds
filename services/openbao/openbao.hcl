ui = false

api_addr = "http://openbao:8200"
cluster_addr = "https://openbao:8201"

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = true
}

storage "raft" {
  path    = "/openbao/file"
  node_id = "openbao-1"
}
