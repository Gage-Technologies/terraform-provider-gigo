data "coder_workspace" "dev" {
}

resource "kubernetes_pod" "dev" {
  count = data.gigo_workspace.dev.transition == "start" ? 1 : 0
}
