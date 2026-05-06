<template>
  <div id="app">
    <router-view v-slot="{ Component }">
      <transition name="page" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  /* 选用更现代的字体 */
  font-family: 'Inter', 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /* 基础背景色，可以更柔和 */
  background-color: #f8f9fa; /* 极浅的灰色 */
}

#app {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
}

router-view {
  flex: 1;
  min-height: 0;
}

/* 页面过渡动画，保留原有的，但可以微调时间 */
.page-enter-active,
.page-leave-active {
  transition: all 0.4s cubic-bezier(0.25, 0.8, 0.25, 1); /* 更流畅的曲线 */
}

.page-enter-from {
  opacity: 0;
  transform: translateX(20px); /* 稍微减小位移 */
}

.page-leave-to {
  opacity: 0;
  transform: translateX(-20px); /* 稍微减小位移 */
}

/* 响应式设计调整 */
@media (max-width: 768px) {
  .page-enter-from,
  .page-leave-to {
    transform: translateX(0); /* 移动端取消水平位移 */
    opacity: 0;
  }
  .page-enter-active,
  .page-leave-active {
    transition: opacity 0.3s ease; /* 移动端只保留淡入淡出 */
  }
}

/* 统一 Element Plus 组件的卡片/按钮/输入框样式 */
.el-card,
.el-button,
.el-input,
.el-message {
  border-radius: 12px !important; /* 统一使用更大的圆角 */
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08) !important; /* 增加卡片质感 */
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1) !important; /* 统一过渡动画 */
}

.el-card:hover,
.el-button:hover:not(:disabled),
.el-input:focus-within {
  transform: translateY(-3px) scale(1.01); /* 增加悬停和聚焦效果 */
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.12) !important;
}

/* 登录/注册页面背景的伪元素，增强高级感 */
.login-container::before,
.register-container::before {
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="40" cy="60" r="1" fill="rgba(255,255,255,0.08)"/><circle cx="60" cy="30" r="1.5" fill="rgba(255,255,255,0.08)"/></svg>') !important;
  opacity: 0.2 !important; /* 降低透明度 */
}

/* 登录/注册卡片，使其更像一个漂浮的卡片 */
.login-card,
.register-card {
  background: rgba(255, 255, 255, 0.92) !important; /* 增加透明度和模糊效果 */
  backdrop-filter: blur(15px) !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15) !important; /* 更强的阴影 */
}

.el-form-item {
  margin-bottom: 24px !important; /* 增加表单项间距 */
}

/* Element Plus v3+ 的输入框样式 */
.el-input__wrapper {
  background: rgba(255, 255, 255, 0.8) !important;
  border-radius: 10px !important;
  box-shadow: none !important; /* 移除内部阴影 */
}
.el-input__wrapper:focus-within {
  transform: scale(1.02);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.5) !important; /* 聚焦时添加光晕 */
  border-color: #409eff !important; /* Add focus border color */
}
/* 针对 Element Plus v2 的可能样式，如果 v3 不适用 */
/*
.el-input__inner {
  border-radius: 10px !important;
  background: rgba(255, 255, 255, 0.8) !important;
  border: none !important;
  box-shadow: none !important;
}
.el-input__inner:focus {
  border-color: #409eff !important;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.5) !important;
}
*/
/* 移除全局滚动条样式，因为它们没有被用到（App.vue 的根级没有滚动条） */
/* ::-webkit-scrollbar { ... } */
/* ::-webkit-scrollbar-track { ... } */
/* ::-webkit-scrollbar-thumb { ... } */
/* ::-webkit-scrollbar-thumb:hover { ... } */
</style>