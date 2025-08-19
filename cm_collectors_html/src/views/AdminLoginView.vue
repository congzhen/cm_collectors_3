<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h2>管理员登录</h2>
        <p>请输入密码访问管理系统</p>
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form"
        @submit.prevent="handleLogin">
        <el-form-item prop="password">
          <el-input v-model="loginForm.password" type="password" placeholder="请输入管理员密码" show-password size="large"
            @keyup.enter="handleLogin">
            <template #prefix>
              <el-icon class="input-icon">
                <Lock />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" :loading="loading" @click="handleLogin" class="login-button"
            native-type="submit">
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>© {{ format(new Date, 'Y') }} CM File Collectors - 文件收集管理系统</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Lock } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useRouter } from 'vue-router'
import { format } from '@/assets/timer'
import { loginServer } from '@/server/login.server'

const router = useRouter()
const loginFormRef = ref<FormInstance>()
const loading = ref(false)

const loginForm = reactive({
  password: ''
})

const loginRules = reactive<FormRules>({
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 1, message: '请输入至少1个字符的密码', trigger: 'blur' }
  ]
})

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        loading.value = true

        // 发送登录请求
        const result = await loginServer.adminLogin(loginForm.password)

        if (result.status) {
          ElMessage.success('登录成功')
          // 登录成功后跳转到主页
          router.push('/')
        } else {
          ElMessage.error(result.msg || '登录失败')
        }
      } catch (error) {
        ElMessage.error('登录过程中发生错误')
        console.error(error)
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 100vw;
  background: linear-gradient(135deg, #1f1f1f 0%, #141422 100%);
  position: relative;
  overflow: hidden;
}

.login-container::before {
  content: "";
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.05) 0%, rgba(255, 255, 255, 0) 70%);
  transform: rotate(30deg);
}

.login-box {
  width: 100%;
  max-width: 400px;
  padding: 40px 30px;
  background: rgba(31, 31, 31, 0.9);
  border-radius: 10px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(10px);
  z-index: 1;
  text-align: center;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.login-header {
  margin-bottom: 30px;
}

.login-header h2 {
  color: #f2f2f2;
  font-size: 24px;
  margin-bottom: 10px;
}

.login-header p {
  color: #aaa;
  font-size: 14px;
}

.login-form {
  margin-bottom: 30px;
}

:deep(.el-input__wrapper) {
  background: rgba(50, 50, 50, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  box-shadow: none;
}

:deep(.el-input__inner) {
  background: transparent;
  color: #f2f2f2;
  border: none;
}

:deep(.el-input__inner::placeholder) {
  color: #888;
}

.input-icon {
  color: #888;
}

:deep(.el-form-item__error) {
  color: #f56c6c;
}

.login-button {
  width: 100%;
  background: linear-gradient(135deg, #409eff 0%, #1a73e8 100%);
  border: none;
  font-weight: bold;
  letter-spacing: 1px;
  margin-top: 10px;
}

.login-button:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
}

.login-footer p {
  color: #888;
  font-size: 12px;
  margin: 0;
}

@media (max-width: 480px) {
  .login-box {
    margin: 0 15px;
    padding: 30px 20px;
  }
}
</style>
