<!-- eslint-disable vue/no-v-html -->
<template>
	<ProjectWrapper
		class="project-overview"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #default>
			<div class="is-max-width-desktop overview-container">
				<Card
					:title="project?.title"
					class="mbe-4"
				>
					<div
						v-if="htmlDescription !== ''"
						class="has-text-start"
						v-html="htmlDescription"
					/>
					<p
						v-else
						class="is-italic has-text-grey"
					>
						{{ $t('project.overview.noDescriptionAvailable') }}
					</p>
				</Card>

				<Card
					:title="$t('project.overview.wiki')"
					class="mbe-4"
				>
					<div class="wiki-placeholder has-text-centered p-4">
						<span class="icon is-large has-text-grey-light mbe-2">
							<Icon
								icon="file"
								size="2x"
							/>
						</span>
						<p class="has-text-grey">
							{{ $t('project.overview.wikiComingSoon') }}
						</p>
					</div>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import DOMPurify from 'dompurify'
import {useI18n} from 'vue-i18n'

import {useProjectStore} from '@/stores/projects'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Card from '@/components/misc/Card.vue'

const props = defineProps<{
	projectId: number,
	isLoadingProject: boolean,
	viewId: number,
}>()

const {t} = useI18n()
const projectStore = useProjectStore()
const project = computed(() => projectStore.projects[props.projectId])

const htmlDescription = computed(() => {
	const description = project.value?.description || ''
	if (description === '') {
		return ''
	}

	if (project.value.id === -1) {
		return t('project.favoriteDescription')
	}

	return DOMPurify.sanitize(description, {ADD_ATTR: ['target']})
})
</script>

<style lang="scss" scoped>
.overview-container {
	margin-inline: auto;
}

.wiki-placeholder {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 2rem;
}
</style>
