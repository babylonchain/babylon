<template>
  <div class="container">
    <div class="row row-sm-revers">
      <div class="col-md-6">
        <SpAssets />
        <SpTokenTransferList />
      </div>
      <div class="col-md-5 col-lg-4 col-md-offset-1 col-lg-offset-2">
        <SpTokenTransfer />
      </div>
    </div>

    <div style="margin-top: 50px">
      <div>Chain Height: {{latestBlock.height}}</div>
      <div>Latest Block:</div>
      <div>{{latestBlock}}</div>
    </div>
  </div>
</template>

<script>
import { SpAssets, SpTokenTransfer, SpTokenTransferList } from '@starport/vue'
import { computed } from 'vue'
import { useStore } from 'vuex'

export default {
  name: 'Portfolio',

  components: { SpTokenTransfer, SpAssets, SpTokenTransferList },

  setup() {
    // store
    let $s = useStore()

    // computed
    let address = computed(() => $s.getters['common/wallet/address'])
    let latestBlock = computed(() => $s.getters['common/blocks/getBlocks'](1)[0])


    return {
      address,
      latestBlock
    }
  }
}
</script>

<style scoped>
.row {
  display: flex;
  flex-wrap: wrap;
}
.col {
  flex-grow: 1;
  padding: 20px;
}
</style>
