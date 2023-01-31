terraform {
  required_providers {
    gigo = {
      source  = "gigo/gigo"
      version = "0.1.0"
    }
    k8s = {
      source = "mingfang/k8s"
    }
  }
}

data "gigo_provisioner" "me" {
}

provider "k8s" {
}

data "gigo_workspace" "me" {
}

resource "gigo_agent" "main" {
  arch = data.gigo_provisioner.me.arch
  os   = data.gigo_provisioner.me.os
}

resource "kubernetes_persistent_volume_claim" "home" {
  metadata {
    name      = "gigo-ws-${data.gigo_workspace.me.owner_id}-${data.gigo_workspace.me.id}-home"
    namespace = "gigo"
  }
  wait_until_bound = false
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = data.gigo_workspace.me.disk
      }
    }
  }
}

# sysbox: namechange
resource "k8s_core_v1_pod" "main" {
  count = data.gigo_workspace.me.start_count
  metadata {
    name      = "gigo-ws-${lower(data.gigo_workspace.me.owner)}-${lower(data.gigo_workspace.me.name)}"
    namespace = "gigo"
    annotations = {
      "io.kubernetes.cri-o.userns-mode" = "auto:size=65536"
    }
  }
  spec {
    runtime_class_name = "sysbox-runc"
    security_context {
      run_asuser = 0
      fsgroup    = 0
    }
    containers {
      name  = "dev"
      image = "codercom/enterprise-base"
      command = ["sh", "-c", <<EOF
      # create user
      echo "Creaing gigo user"
      useradd --create-home --shell /bin/bash gigo

      # initialize the gigo home directory using /etc/skeleton
      cp -r /etc/skel/. /home/gigo/

      # change ownership of gigo directory
      echo "Ensuring directory ownership for gigo user"
      chown gigo:gigo -R /home/gigo

      # disable sudo for gigo user
      echo "gigo ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/gigo

      # Start the Gigo agent as the "gigo" user
      # once systemd has started up
      echo "Waiting for systemd to start"
      sudo -u gigo --preserve-env=CODER_AGENT_TOKEN /bin/bash -- <<-'      EOT' &
      while [[ ! $(systemctl is-system-running) =~ ^(running|degraded) ]]
      do
        echo "Waiting for system to start... $(systemctl is-system-running)"
        sleep 2
      done

      echo "Starting Gigo agent"
      ${gigo_agent.main.init_script}
      EOT

      echo "Executing /sbin/init"
      exec /sbin/init

      echo "Exiting"
      EOF
      ]

      env {
        name  = "GIGO_AGENT_TOKEN"
        value = gigo_agent.main.token
      }
      volume_mounts {
        mount_path = "/home/gigo"
        name       = "home"
        read_only  = false
      }

      resources {
        requests = {
          cpu    = "500m"
          memory = "500Mi"
        }
        limits = {
          cpu    = data.gigo_workspace.me.cpu
          memory = data.gigo_workspace.me.mem
        }
      }
    }

    volumes {
      name = "home"
      persistent_volume_claim {
        claim_name = kubernetes_persistent_volume_claim.home.metadata.0.name
        read_only  = false
      }
    }
  }
}