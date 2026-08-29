ui = false

api_addr = "http://openbao:8200"
cluster_addr = "https://openbao:8201"

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = true
}

seal "static" {
  current_key_id = "bloodforbuds-static-seal-v1"
  current_key    = "env://OPENBAO_UNSEAL_KEY"
}

storage "raft" {
  path    = "/openbao/file"
  node_id = "openbao-1"
}
