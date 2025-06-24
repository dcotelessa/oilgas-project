<template>
  <div id="app">
    <header>
      <h1>Oil & Gas Inventory System</h1>
    </header>
    
    <nav>
      <router-link to="/">Dashboard</router-link>
      <router-link to="/customers">Customers</router-link>
      <router-link to="/inventory">Inventory</router-link>
      <router-link to="/grades">Grades</router-link>
    </nav>
    
    <main>
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
// Vue 3 Composition API setup
</script>

<style scoped>
header {
  background: #2c3e50;
  color: white;
  padding: 1rem;
  text-align: center;
}

nav {
  background: #34495e;
  padding: 1rem;
}

nav a {
  color: white;
  text-decoration: none;
  margin-right: 1rem;
  padding: 0.5rem;
}

nav a:hover {
  background: #4a5f7a;
  border-radius: 4px;
}

main {
  padding: 2rem;
}
</style>
