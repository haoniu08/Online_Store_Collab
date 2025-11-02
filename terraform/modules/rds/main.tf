resource "aws_db_subnet_group" "this" {
  name       = lower("${var.service_name}-db-subnets")
  subnet_ids = var.subnet_ids

  tags = {
    Name = "${var.service_name}-db-subnets"
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.service_name}-rds-sg"
  description = "RDS security group allowing MySQL from ECS"
  vpc_id      = var.vpc_id

  ingress {
    description = "MySQL from ECS"
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    security_groups = var.allowed_security_group_ids
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "this" {
  identifier              = lower("${var.service_name}-mysql")
  engine                  = "mysql"
  engine_version          = "8.0"
  instance_class          = "db.t3.micro"
  allocated_storage       = 20
  db_name                 = var.db_name
  username                = var.db_username
  password                = random_password.db_password.result
  port                    = 3306
  publicly_accessible     = false
  vpc_security_group_ids  = [aws_security_group.rds.id]
  db_subnet_group_name    = aws_db_subnet_group.this.name
  skip_final_snapshot     = true
  deletion_protection     = false

  # For classroom assignments; not for production
  backup_retention_period = 0

  tags = {
    Name = "${var.service_name}-mysql"
  }
}

resource "random_password" "db_password" {
  length  = 20
  special = true
}


