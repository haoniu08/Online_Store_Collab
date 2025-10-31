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

# SNS Topic for order processing events (Homework 7)
module "sns" {
  source       = "./modules/sns"
  service_name = var.service_name
  environment  = "dev"
}

# SQS Queue for order processing (Homework 7)
module "sqs" {
  source        = "./modules/sqs"
  service_name  = var.service_name
  environment   = "dev"
  sns_topic_arn = module.sns.topic_arn
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

# RDS MySQL (Homework 8)
module "rds" {
  source                      = "./modules/rds"
  service_name                = var.service_name
  vpc_id                      = module.network.vpc_id
  subnet_ids                  = module.network.subnet_ids
  allowed_security_group_ids  = [module.network.ecs_security_group_id]
  db_name                     = var.db_name
  db_username                 = var.db_username
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:api-server"
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

  # Pass SNS and SQS info to containers (Homework 7)
  sns_topic_arn      = module.sns.topic_arn
  sqs_queue_url      = module.sqs.queue_url
}

# Order Processor ECS Service (Homework 7)
module "ecs_processor" {
  source             = "./modules/ecs_processor"
  service_name       = var.service_name
  cluster_id         = module.ecs.cluster_id
  image              = "${module.ecr.repository_url}:processor"
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  desired_count      = 1  # Start with 1 task as per assignment
  region             = var.aws_region
  cpu                = var.cpu
  memory             = var.memory
  sqs_queue_url      = module.sqs.queue_url
  worker_count       = 100  # Phase 5: Testing with 100 worker goroutines (assignment maximum)
}


// Build & push the API server image into ECR
resource "docker_image" "app" {
  name = "${module.ecr.repository_url}:api-server"

  build {
    context = ".."
    dockerfile = "../Dockerfile"
  }
}

resource "docker_registry_image" "app" {
  name = docker_image.app.name
}

// Build & push the order processor image into ECR
resource "docker_image" "processor" {
  name = "${module.ecr.repository_url}:processor"

  build {
    context = ".."
    dockerfile = "../Dockerfile.processor"
  }
}

resource "docker_registry_image" "processor" {
  name = docker_image.processor.name
}
