# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:circuit-breaker"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
  cpu                = var.cpu
  memory             = var.memory
}


// Build & push the Go app image with circuit breaker into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "circuit-breaker"
  name = "${module.ecr.repository_url}:circuit-breaker"

  build {
    # relative path from terraform/ to project root
    context = ".."
    # Dockerfile is in the project root
  }
}

resource "docker_registry_image" "app" {
  # this will push :circuit-breaker → ECR
  name = docker_image.app.name
}
