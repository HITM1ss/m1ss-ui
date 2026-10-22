<template>
  <ServiceVue
    :id="modal.id"
    v-model="modal.visible"
    :visible="modal.visible"
    :data="modal.data"
    :in-tags="inTags"
    :ts-tags="tsTags"
    :ss-tags="ssTags"
    :tls-configs="tlsConfigs"
    @close="closeModal"
  />
  <div class="panel-toolbar">
    <h1 class="panel-toolbar__title">
      {{ $t('pages.services') }}
    </h1>
    <v-btn
      color="primary"
      @click="showModal(0)"
    >
      {{ $t('actions.add') }}
    </v-btn>
  </div>
  <div
    v-if="services.length === 0"
    class="panel-empty"
  >
    <v-icon
      icon="mdi-server-outline"
      size="32"
    />
    {{ $t('noData') }}
  </div>
  <v-row
    v-else
    class="resource-grid"
  >
    <v-col
      v-for="(item, index) in <any[]>services"
      :key="item.tag"
    >
      <v-card
        class="resource-card"
        :title="item.tag"
      >
        <v-card-subtitle>
          <v-row>
            <v-col>{{ item.type }}</v-col>
          </v-row>
        </v-card-subtitle>
        <v-card-text>
          <v-row>
            <v-col>{{ $t('in.addr') }}</v-col>
            <v-col>
              {{ item.listen }}
            </v-col>
          </v-row>
          <v-row>
            <v-col>{{ $t('in.port') }}</v-col>
            <v-col>
              {{ item.listen_port }}
            </v-col>
          </v-row>
          <v-row>
            <v-col>{{ $t('objects.tls') }}</v-col>
            <v-col>
              {{ item.tls_id > 0 ? $t('enable') : $t('disable') }}
            </v-col>
          </v-row>
        </v-card-text>
        <v-divider />
        <v-card-actions>
          <v-btn
            icon="mdi-file-edit"
            :aria-label="$t('actions.edit')"
            @click="showModal(item.id)"
          >
            <v-icon />
            <v-tooltip
              activator="parent"
              location="top"
              :text="$t('actions.edit')"
            />
          </v-btn>
          <v-btn
            icon="mdi-file-remove"
            :aria-label="$t('actions.del')"
            style="margin-inline-start:0;"
            color="warning"
            @click="delOverlay[index] = true"
          >
            <v-icon />
            <v-tooltip
              activator="parent"
              location="top"
              :text="$t('actions.del')"
            />
          </v-btn>
          <v-overlay
            v-model="delOverlay[index]"
            contained
            class="align-center justify-center"
          >
            <v-card
              :title="$t('actions.del')"
              rounded="lg"
            >
              <v-divider />
              <v-card-text>{{ $t('confirm') }}</v-card-text>
              <v-card-actions>
                <v-btn
                  color="error"
                  variant="outlined"
                  @click="delSrv(item.id)"
                >
                  {{ $t('yes') }}
                </v-btn>
                <v-btn
                  color="success"
                  variant="outlined"
                  @click="delOverlay[index] = false"
                >
                  {{ $t('no') }}
                </v-btn>
              </v-card-actions>
            </v-card>
          </v-overlay>
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
import Data from '@/store/modules/data'
import { Srv } from '@/types/services'
import { tls } from '@/types/tls'
import { computed, ref } from 'vue'
import ServiceVue from '@/layouts/modals/Service.vue'

const services = computed((): Srv[] => {
  return <Srv[]> Data().services
})

const tsTags = computed((): string[] => {
  return Data().endpoints?.filter(o => o.type == "tailscale")?.map(o => o.tag)
})

// Only the inbound types that declare users are in the union, but any inbound
// can carry one, so the field is read through a shape that only says that.
const ssTags = computed((): string[] => {
  return Data().inbounds?.filter(o => o.type == "shadowsocks" && !(o as { users?: unknown }).users)?.map(o => o.tag)
})

const inTags = computed((): string[] => {
  return [...(Data().inbounds?.map(o => o.tag).filter(t => t != null) ?? []), ...(Data().endpoints?.filter(e => e.listen_port > 0).map(e => e.tag) ?? [])]
})

const tlsConfigs = computed((): tls[] => {
  return Data().tlsConfigs
})

const modal = ref({
  visible: false,
  id: 0,
  data: "",
})

let delOverlay = ref(new Array<boolean>)

const showModal = (id: number) => {
  modal.value.id = id
  modal.value.data = id == 0 ? '' : JSON.stringify(services.value.findLast(o => o.id == id))
  modal.value.visible = true
}

const closeModal = () => {
  modal.value.visible = false
}

const delSrv = async (id: number) => {
  const index = services.value.findIndex(i => i.id == id)
  const tag = services.value[index].tag

  const success = await Data().save("services", "del", tag)
  if (success) delOverlay.value[index] = false
}
</script>
