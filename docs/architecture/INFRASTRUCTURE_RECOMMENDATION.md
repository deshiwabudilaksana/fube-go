# Fube-Go Infrastructure Recommendations

**Date:** March 13, 2026
**Project Type:** Go Monolith with HTMX Frontend and PostgreSQL Database (Post-Pivot)

---

## 1. Project Profile & Needs
The `fube-go` project is an incredibly efficient stack. A Go binary requires very little RAM and CPU compared to Node.js or Ruby, and serving HTML directly via HTMX means there is no heavy client-side application to host on a CDN.

**Key Requirements:**
- A place to run a single Docker container (the Go binary).
- A managed PostgreSQL database (supporting JSONB).
- Zero downtime deployments.
- Minimal DevOps overhead ("YAML engineering").

---

## 2. The Recommended Stack: The "Modern PaaS"

For a project of this architectural profile, a modern Platform-as-a-Service (PaaS) is the absolute best choice. It provides the perfect balance between control, cost, and developer experience.

### A. Compute / Hosting: Railway, Render, or Fly.io
Instead of managing a raw Linux server (VPS) or dealing with the complexity of Kubernetes (AWS EKS), deploy your Docker container to a modern PaaS.

*   **Why Railway / Render / Fly.io?**
    *   **Deploy from Git:** Push to the `main` branch, and the platform automatically builds your Dockerfile and deploys the new container with zero downtime.
    *   **Built-in SSL & Routing:** You don't need to configure Nginx, Traefik, or Let's Encrypt. It handles HTTPS automatically.
    *   **Autoscaling:** They can automatically spin up more instances of your Go app if traffic spikes.

### B. Database: Managed PostgreSQL 16+
Since we are pivoting to a "Power of One" database strategy (dropping MongoDB in favor of Postgres JSONB), the database is the most critical piece of infrastructure.

*   **Recommendation:** Use the Managed PostgreSQL add-on provided by your chosen PaaS (e.g., Railway Postgres or Render Postgres), or a dedicated serverless provider like **Neon** or **Supabase**.
*   **Why?**
    *   Automated daily backups.
    *   Point-in-time recovery (PITR).
    *   No need to manage database updates, disk space scaling, or tuning `postgresql.conf`.

### C. Containerization: Docker
Wrap your Go application in a multi-stage `Dockerfile`.
*   **Stage 1:** Use the official `golang` image to compile the binary.
*   **Stage 2:** Use `alpine` or `scratch` (distroless) to run the binary.
*   **Why?** This results in an incredibly small container image (often < 20MB) that starts in milliseconds and is highly secure since it contains no OS utilities.

### D. CI/CD Pipeline: GitHub Actions
Use GitHub Actions to automate your testing and deployment.
*   **Step 1:** Run `go test` and `go vet` on every Pull Request.
*   **Step 2:** On merge to `main`, trigger the PaaS deployment webhook or push the built Docker image to their registry.

---

## 3. Alternative Infrastructure Options (And Why to Avoid Them)

### Option A: The "Bare Metal" VPS (DigitalOcean Droplet, AWS EC2)
*   **Setup:** You rent a Linux server, install Docker, Nginx, Let's Encrypt, and Postgres yourself.
*   **Pros:** It is the absolute cheapest option per month.
*   **Cons:** You are now a part-time SysAdmin. You have to handle OS security updates, manual database backups, configuring firewall rules, and setting up deployment scripts (like Capistrano or custom GitHub Actions via SSH).
*   **Verdict:** Not recommended unless you have zero budget and abundant free time.

### Option B: "Enterprise Cloud" (AWS ECS/EKS, Google Kubernetes Engine)
*   **Setup:** Deploying Docker containers to a Kubernetes cluster or AWS Fargate.
*   **Pros:** Infinite scalability, highly available, industry standard for massive teams.
*   **Cons:** Extremely complex. You will spend more time writing Terraform and Kubernetes YAML files than writing Go code. It is also significantly more expensive due to control-plane fees and NAT Gateways.
*   **Verdict:** Massive overkill for a Monolith. Avoid until the engineering team grows significantly and the app requires microservices.

---

## 4. Free Tier Options (Zero Cost Setup)

If you are looking to run this project entirely for free (ideal for a side project or early-stage MVP), here is how you can combine free tiers from top providers:

### Compute (Free Hosting for Go)
*   **Render:** Offers a free web service tier. It spins down after 15 minutes of inactivity (causing a cold start delay on the next request), but it is perfect for development or low-traffic side projects.
*   **Fly.io:** Offers up to 3 shared-cpu-1x VMs (256MB RAM) for free. Because Go binaries use very little memory, 256MB is more than enough. **(Recommended)**

### Database Option 1: Full PostgreSQL (The "Power of One" Pivot)
*   **Neon.tech:** Offers a generous free tier (0.5 GiB storage, serverless compute) and provides you with a direct PostgreSQL connection string. You are already using Neon in your `config.go`, so this is a seamless fit!
*   **Supabase:** Offers a very robust free tier for PostgreSQL (500MB storage, 2 active projects).

### Database Option 2: Full MongoDB (Retaining Current Architecture)
If you prefer to keep your current dual-database setup, or if you decide to pivot to a **Full MongoDB** architecture (dropping PostgreSQL entirely), you can easily run MongoDB for free:
*   **MongoDB Atlas (M0 Free Cluster):** Atlas provides an exceptional free tier that gives you 512MB of storage on a shared cluster (AWS, Google Cloud, or Azure). 
    *   **Pros:** It never spins down, it includes automatic minor-version patches, and it offers built-in visual data browsing.
    *   **Cons:** Limited to 500 connections, and network performance can occasionally spike because it is a shared tenant environment.

**The Ultimate Free Stack:** Host your Go binary on **Fly.io** (always-on, no sleep). If you pivot to Postgres, use **Neon**. If you choose to go Full MongoDB (or maintain the dual setup), use **MongoDB Atlas**. This gives you a fast, production-like environment for $0/month regardless of your database choice.

---

## 5. Summary Verdict

The **"Docker + Modern PaaS (Railway/Render/Fly.io) + Managed Postgres"** stack is the gold standard for Go monoliths today. It allows you to focus 100% of your energy on writing Go and HTMX features, while the platform handles the boring operational work for a very reasonable monthly cost (or even for free using the stack above!).