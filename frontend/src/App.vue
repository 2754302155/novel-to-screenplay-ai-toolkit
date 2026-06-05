<template>
  <div class="app-shell">
    <aside class="sidebar">
      <div>
        <p class="eyebrow">Novel to Screenplay</p>
        <h1>AI 小说转剧本工具</h1>
      </div>
      <nav aria-label="主导航">
        <RouterLink to="/">工作台</RouterLink>
        <RouterLink to="/tasks">转换任务</RouterLink>
      </nav>
    </aside>
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>
