# Wire together modules: network, ecr, logging, alb, ecs with auto scaling

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

# Application Load Balancer for horizontal scaling
module "alb" {
  source             = "./modules/alb"
  service_name       = var.service_name
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.alb_security_group_id]
  vpc_id             = module.network.vpc_id
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.min_instances
  region             = var.aws_region
  cpu                = var.cpu
  memory             = var.memory
  target_group_arn   = module.alb.target_group_arn
  alb_listener_arn   = module.alb.listener_arn
  min_capacity       = var.min_instances
  max_capacity       = var.max_instances
  cpu_target_percentage = var.cpu_target_percentage
}


// Build & push the Go app image into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "latest"
  name = "${module.ecr.repository_url}:latest"

  build {
    # relative path from terraform/ to project root
    context = ".."
    # Dockerfile is in the project root
  }
}

resource "docker_registry_image" "app" {
  # this will push :latest → ECR
  name = docker_image.app.name
}
