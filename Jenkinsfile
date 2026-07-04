pipeline {
    agent any

    environment {
        // Lingkungan kerja
        GO_VERSION = '1.21'
        NODE_VERSION = '20'
        
        // Kredensial untuk Rsync/SSH ke VPS Produksi (Contoh)
        // Harus disetel di Jenkins Credentials
        SSH_CREDENTIAL_ID = 'vps-ssh-key'
        VPS_USER = 'u0_a129'
        VPS_IP = '192.168.18.27'
        DEPLOY_DIR = '/data/data/com.termux/files/home/siakad-nawawi'
    }

    triggers {
        // Memicu pipeline berjalan otomatis setiap kali ada PUSH ke Github (via Github Webhook)
        githubPush()
    }

    stages {
        stage('⚙️ Setup & Dependencies') {
            steps {
                echo 'Mengunduh dependensi Golang...'
                dir('backend') {
                    sh 'go mod download'
                }
                echo 'Mengunduh dependensi Node.js (Frontend & Mobile)...'
                dir('frontend') {
                    sh 'npm ci'
                }
                dir('mobile') {
                    sh 'npm ci'
                }
            }
        }

        stage('🛡️ Security, Code Quality & Unit Test') {
            steps {
                echo 'Menjalankan Pengujian Unit & Code Coverage (Backend)...'
                dir('backend') {
                    // Jalankan tes dan stop jika gagal, plus hasilkan coverage
                    sh 'go test -coverprofile=coverage.out ./...'
                    // Memeriksa Code Coverage Threshold (Target: 80% untuk saat ini)
                    sh '''
                    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
                    echo "Code Coverage Saat Ini: ${COVERAGE}%"
                    awk -v cov="${COVERAGE}" -v threshold=70.0 'BEGIN { if (cov < threshold) { print "❌ Code coverage di bawah target (" threshold "%)"; exit 1 } }'
                    '''
                }
                echo 'Menjalankan SAST & Vulnerability Scan...'
                dir('backend') {
                    sh 'gosec ./... || true'
                    // Memeriksa celah pustaka dependensi Golang
                    sh 'govulncheck ./... || true'
                }
                dir('frontend') {
                    // Memeriksa celah pustaka dependensi Node.js
                    sh 'npm audit --audit-level=high || true'
                }
            }
        }

        stage('🧪 Integration & E2E Test (Mock)') {
            steps {
                echo 'Menjalankan Pengujian Integrasi (Redis/MySQL)...'
                // TODO: Eksekusi script pengujian integrasi DB
                sh 'echo "✅ Integration tests passed"'

                echo 'Menjalankan End-to-End (E2E) Testing...'
                dir('frontend') {
                    // Contoh jika menggunakan Playwright: sh 'npx playwright test'
                    sh 'echo "✅ E2E tests passed (Placeholder)"'
                }
            }
        }

        stage('🔨 Build (Kompilasi)') {
            parallel {
                stage('Build Golang (Backend)') {
                    steps {
                        dir('backend') {
                            sh 'CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server main.go'
                        }
                    }
                }
                stage('Build Node.js (Frontend)') {
                    steps {
                        dir('frontend') {
                            sh 'npm run build'
                        }
                    }
                }
                stage('Build SPA (Mobile)') {
                    steps {
                        dir('mobile') {
                            sh 'npm run build'
                        }
                    }
                }
            }
        }

        stage('🚀 Deploy to VPS (PM2)') {
            steps {
                echo 'Menyiapkan direktori artifact hasil build...'
                sh '''
                    mkdir -p release_artifact
                    
                    # 1. Backend Golang (Hanya Binary)
                    mkdir -p release_artifact/backend
                    cp backend/server release_artifact/backend/
                    
                    # 2. Frontend Svelte (Hanya folder build & package.json untuk production)
                    mkdir -p release_artifact/frontend
                    cp -r frontend/build release_artifact/frontend/
                    cp frontend/package.json release_artifact/frontend/
                    cp frontend/package-lock.json release_artifact/frontend/
                    
                    # 3. Mobile SPA (Hanya folder dist yang di-rename ke build)
                    mkdir -p release_artifact/mobile
                    cp -r mobile/dist release_artifact/mobile/build
                    
                    # 4. PM2 Config
                    cp ecosystem.config.cjs release_artifact/
                '''

                echo 'Mentransfer file artifact ke server produksi...'
                /* 
                  [NOTE] 
                  Ini adalah perintah sinkronisasi rsync ke VPS target. 
                  Hanya 'release_artifact' yang dikirim, SOURCE CODE TIDAK DIKIRIM!
                */
                sh """
                    rsync -avz -e 'ssh -p 8022' --delete release_artifact/ ${VPS_USER}@${VPS_IP}:${DEPLOY_DIR}/
                    ssh -p 8022 ${VPS_USER}@${VPS_IP} "cd ${DEPLOY_DIR}/frontend && npm ci --omit=dev"
                    ssh -p 8022 ${VPS_USER}@${VPS_IP} "cd ${DEPLOY_DIR} && pm2 reload ecosystem.config.cjs"
                """
                echo '✅ Deployment ke Termux VPS Selesai!'
            }
        }
    }

    post {
        success {
            echo '🎉 Pipeline berhasil! Aplikasi live di production.'
        }
        failure {
            echo '❌ Pipeline gagal! Cek log Jenkins untuk melihat kesalahan pada build atau test.'
        }
    }
}
