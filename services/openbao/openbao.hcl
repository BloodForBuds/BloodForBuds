ui = false

api_addr = "http://openbao:8200"
cluster_addr = "https://openbao:8201"

listener "tcp" {
  address                                  = "0.0.0.0:8200"
  disable_unauthed_generate_root_endpoints = true
  disable_unauthed_rekey_endpoints         = true
  tls_disable                              = true
}

seal "static" {
  current_key_id = "bloodforbuds-static-seal-v1"
  current_key    = "env://OPENBAO_UNSEAL_KEY"
}

storage "raft" {
  path    = "/openbao/file"
  node_id = "openbao-1"
}

initialize "access" {
  request "create-backup-policy" {
    operation = "update"
    path      = "sys/policies/acl/bloodforbuds-backup"
    data = {
      policy = {
        eval_source = "file"
        eval_type   = "string"
        path        = "/openbao/backup-policy.hcl"
      }
    }
  }

  request "create-operator-policy" {
    operation = "update"
    path      = "sys/policies/acl/bloodforbuds-operator"
    data = {
      policy = {
        eval_source = "file"
        eval_type   = "string"
        path        = "/openbao/operator-policy.hcl"
      }
    }
  }

  request "enable-approle" {
    operation = "update"
    path      = "sys/auth/approle"
    data = {
      type        = "approle"
      description = "Machine authentication for BloodForBuds"
    }
  }

  request "create-backup-role" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-backup"
    data = {
      bind_secret_id          = true
      secret_id_num_uses      = 0
      secret_id_ttl           = 0
      token_max_ttl           = "30m"
      token_no_default_policy = true
      token_num_uses          = 0
      token_policies          = ["bloodforbuds-backup"]
      token_ttl               = "30m"
      token_type              = "service"
    }
  }

  request "set-backup-role-id" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-backup/role-id"
    data = {
      role_id = {
        eval_source     = "env"
        eval_type       = "string"
        env_var         = "BAO_BACKUP_ROLE_ID"
        require_present = true
      }
    }
  }

  request "register-backup-secret-id" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-backup/custom-secret-id"
    data = {
      secret_id = {
        eval_source     = "env"
        eval_type       = "string"
        env_var         = "BAO_BACKUP_SECRET_ID"
        require_present = true
      }
    }
  }

  request "create-operator-role" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-operator"
    data = {
      bind_secret_id          = true
      secret_id_num_uses      = 0
      secret_id_ttl           = 0
      token_max_ttl           = "30m"
      token_no_default_policy = true
      token_num_uses          = 0
      token_policies          = ["bloodforbuds-operator"]
      token_ttl               = "15m"
      token_type              = "service"
    }
  }

  request "set-operator-role-id" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-operator/role-id"
    data = {
      role_id = {
        eval_source     = "env"
        eval_type       = "string"
        env_var         = "BAO_OPERATOR_ROLE_ID"
        require_present = true
      }
    }
  }

  request "register-operator-secret-id" {
    operation = "update"
    path      = "auth/approle/role/bloodforbuds-operator/custom-secret-id"
    data = {
      secret_id = {
        eval_source     = "env"
        eval_type       = "string"
        env_var         = "BAO_OPERATOR_SECRET_ID"
        require_present = true
      }
    }
  }
}
