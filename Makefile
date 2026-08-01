# ==========================================
# Konfigurasi Termux (Staging)
# ==========================================
STAGING_USER=u0_a129
STAGING_IP=100.77.99.42
STAGING_PORT=8022
STAGING_DIR=~/siakad-nawawi

# ==========================================
# Konfigurasi IDCloudHost (Production)
# ==========================================
PROD_USER=iqbalmahad4039
PROD_IP=210.79.190.11
PROD_PORT=22
PROD_DIR=/var/www/siakad-nawawi-be-fe-terbaru

.PHONY: deploy-staging deploy-prod

# ------------------------------------------
# Perintah: make deploy-staging
# ------------------------------------------
deploy-staging:
	@echo "🚀 Memulai Deployment ke Termux (Staging)..."
	@echo "🔐 Mengirim file rahasia (.env) dengan aman..."
	scp -P $(STAGING_PORT) .env.staging $(STAGING_USER)@$(STAGING_IP):$(STAGING_DIR)/.env.staging
	scp -P $(STAGING_PORT) frontend/.env.staging $(STAGING_USER)@$(STAGING_IP):$(STAGING_DIR)/frontend/.env.staging
	ssh -p $(STAGING_PORT) $(STAGING_USER)@$(STAGING_IP) "\
		cd $(STAGING_DIR) && \
		git pull origin main && \
		cp .env.staging .env && \
		echo '📦 Membangun Backend Golang...' && \
		cd backend && go build -o server && cd .. && \
		echo '🎨 Membangun Frontend SvelteKit...' && \
		cd frontend && cp .env.staging .env.production.local && npm install && npm run build && cd .. && \
		echo '🔄 Merestart PM2...' && \
		pm2 restart all"
	@echo "✅ Deploy ke Staging Selesai!"

# ------------------------------------------
# Perintah: make deploy-prod
# ------------------------------------------
deploy-prod:
	@echo "🔥 Memulai Deployment ke IDCloudHost (Production)..."
	@echo "🔐 Mengirim file rahasia (.env) dengan aman..."
	scp -P $(PROD_PORT) .env.production $(PROD_USER)@$(PROD_IP):$(PROD_DIR)/.env.production
	scp -P $(PROD_PORT) frontend/.env.production $(PROD_USER)@$(PROD_IP):$(PROD_DIR)/frontend/.env.production
	ssh -p $(PROD_PORT) $(PROD_USER)@$(PROD_IP) "\
		cd $(PROD_DIR) && \
		git pull origin main && \
		cp .env.production .env && \
		echo '📦 Membangun Backend Golang...' && \
		cd backend && go build -o server && cd .. && \
		echo '🎨 Membangun Frontend SvelteKit...' && \
		cd frontend && cp .env.production .env.production.local && npm install && npm run build && cd .. && \
		echo '🔄 Merestart PM2...' && \
		pm2 restart all"
	@echo "✅ Deploy ke Production Selesai!"
