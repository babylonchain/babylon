<template>
  <div>
    <nav class="navbar navbar-expand-lg bg-light">
      <div class="container-fluid">
        <a class="navbar-brand" href="#">Navbar</a>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNavAltMarkup" aria-controls="navbarNavAltMarkup" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNavAltMarkup">
          <div class="navbar-nav">
            <a class="nav-link active" aria-current="page" href="/">Home</a>
            <a class="nav-link active" aria-current="page" href="/portfolio">Portfolio</a>
            <a class="nav-link active" aria-current="page" href="/data">Data</a>
          </div>
          <!-- <form class="d-flex" role="search">
            <input class="form-control me-2" type="search" placeholder="Search" aria-label="Search">
            <button class="btn btn-outline-success" type="submit">Search</button>
          </form> -->
        </div>
      </div>
    </nav>

    <router-view style="margin-top: 50px"/>
  </div>
</template>

<script lang="ts">
import { SpNavbar, SpTheme } from '@starport/vue'
import { computed, onBeforeMount } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from 'vuex'

export default {
  components: { SpTheme, SpNavbar },

  setup() {
    // store
    let $s = useStore()

    // router
    let router = useRouter()

    // state
    let navbarLinks = [
      { name: 'Portfolio', url: '/portfolio' },
      { name: 'Data', url: '/data' }
    ]

    // computed
    let address = computed(() => $s.getters['common/wallet/address'])
    // lh
    onBeforeMount(async () => {
      await $s.dispatch('common/env/init')
      // router.push('portfolio')
    })

    return {
      navbarLinks,
      // router
      router,
      // computed
      address,
    }
  }
}
</script>

<style scoped lang="scss">
  body {
    margin: 0;
  }
  .navbar a{
      font-size: large;
  }
</style>
