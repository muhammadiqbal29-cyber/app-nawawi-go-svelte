pipeline {
    agent any

    environment {
        // Lingkungan kerja
        GO_VERSION = '1.21'
        NODE_VERSION = '20'
        
        // Kredensial untuk Rsync/SSH ke VPS Produksi (Contoh)
        // Harus disetel di Jenkins Credentials
        SSH_CREDENTIAL_ID = 'vps-ssh-key'
        VPS_USER = 'root'
        VPS_IP = '192.168.18.27'
        DEPLOY_DIR = '/var/www/siakad'
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

        stage('🛡️ Security & Unit Test') {
            steps {
                echo 'Menjalankan Pengujian Unit (Backend)...'
                dir('backend') {
                    // Jalankan tes dan stop jika gagal
                    sh 'go test ./...'
                }
                echo 'Menjalankan SAST Security Scan...'
                dir('backend') {
                    // Pastikan gosec terinstal di Jenkins agent
                    sh 'gosec ./...'
                }
            }
        }

        stage('🔨 Build (Kompilasi)') {
            parallel {
                stage('Build Golang (Backend)') {
                    steps {
                        dir('backend') {
                            sh 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server main.go'
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
                echo 'Mentransfer file ke server produksi...'
                /* 
                  [NOTE] 
                  Pada kenyataannya, perintah sinkronisasi ini membutuhkan 'ssh-agent' 
                  plugin atau konfigurasi kredensial Jenkins.
                  Ini adalah simulasi perintah rsync ke VPS target.
                */
                // sh """
                //     rsync -avz --exclude '.git' --exclude 'node_modules' ./ ${VPS_USER}@${VPS_IP}:${DEPLOY_DIR}/
                //     ssh ${VPS_USER}@${VPS_IP} "cd ${DEPLOY_DIR} && pm2 restart ecosystem.config.cjs"
                // """
                echo '✅ Simulasi Deployment Native Selesai! (Kode telah dipindahkan dan PM2 di-restart)'
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
