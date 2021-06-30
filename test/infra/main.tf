terraform {
  required_providers {
    metal = {
      source = "equinix/metal"
      # version = "1.0.0"
    }
  }
}

variable "auth_token" {
  type = string
}

variable "ssh_port" {
  description = "The port the EC2 Instance should listen on for SSH requests."
  type        = number
  default     = 22
}

variable "ssh_user" {
  description = "SSH user name to use for remote exec connections,"
  type        = string
  default     = "root"
}

provider "metal" {
  auth_token = var.auth_token
}

resource "metal_project" "quicksilver_e2e" {
   name = "quicksilver_e2e"
}

resource "metal_project_ssh_key" "test" {
  name       = "e2e"
  public_key = file("./.ssh/id_rsa.pub")
  project_id = metal_project.quicksilver_e2e.id
}

resource "metal_device" "web1" {
  hostname         = "web1"
  plan             = "c1.small.x86"
  facilities       = ["ewr1"]
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = metal_project.quicksilver_e2e.id
}

resource "null_resource" "example_provisioner" {
    triggers = {
      "public_ip" = "metal_device.web1.network.0.address"
    }

    connection {
        type = "ssh"
        host = metal_device.web1.network.0.address
        user = var.ssh_user
        port = var.ssh_port
        private_key = file("./.ssh/id_rsa")
        agent = true
    }

    provisioner "file" {
        source      = "files/get-meta.sh"
        destination = "/tmp/get-meta.sh"
    }

    provisioner "remote-exec" {
        inline = [
            "chmod +x /tmp/get-meta.sh",
            "/tmp/get-meta.sh > /tmp/metadata",
        ]
    }

    # provisioner "local-exec" {
    #     # copy the metadata file back to CWD, which will be tested
    #     command = "scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${var.ssh_user}@${metal_device.web1.network.0}:/tmp/metadata metadata"
    # }
}