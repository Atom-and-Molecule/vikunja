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
				<!-- Project Description Card -->
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

				<!-- Wiki Home Page Card -->
				<Card
					v-if="homePage"
					class="mbe-4"
				>
					<header class="is-flex is-justify-content-between is-align-items-center mbe-3">
						<div class="is-flex is-align-items-center gap-2">
							<h2 class="title is-4 mb-0">
								{{ homePage.title }}
							</h2>
							<span class="tag is-info is-light is-small">
								{{ $t('project.wiki.homePage') }}
							</span>
						</div>

						<div class="buttons is-right mb-0">
							<BaseButton
								class="is-small is-light"
								:to="{ name: 'project.wiki.page', params: { projectId, pageId: homePage.id } }"
							>
								<span>{{ $t('project.wiki.allPages', { count: pages.length }) }}</span>
							</BaseButton>
							<BaseButton
								v-if="canWrite"
								class="is-small is-primary"
								:to="{ name: 'project.wiki.page', params: { projectId, pageId: homePage.id }, query: { edit: 'true' } }"
							>
								<span class="icon is-small">
									<Icon icon="pen" />
								</span>
								<span>{{ $t('project.wiki.editPage') }}</span>
							</BaseButton>
						</div>
					</header>

					<div
						v-if="htmlWikiContent !== ''"
						class="content has-text-start"
						v-html="htmlWikiContent"
					/>
					<p
						v-else
						class="is-italic has-text-grey"
					>
						{{ $t('project.overview.wikiComingSoon') }}
					</p>
				</Card>

				<!-- Empty State for Wiki -->
				<Card
					v-else
					:title="$t('project.overview.wiki')"
					class="mbe-4"
				>
					<div class="wiki-placeholder has-text-centered p-5">
						<span class="icon is-large has-text-grey-light mbe-2">
							<Icon
								icon="file"
								size="2x"
							/>
						</span>
						<p class="has-text-grey mbe-3">
							{{ $t('project.wiki.noPages') }}
						</p>
						<BaseButton
							v-if="canWrite"
							class="button is-primary is-small"
							:to="{ name: 'project.wiki', params: { projectId }, query: { new: 'true' } }"
						>
							<span class="icon is-small">
								<Icon icon="plus" />
							</span>
							<span>{{ $t('project.wiki.createHomePage') }}</span>
						</BaseButton>
					</div>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import DOMPurify from 'dompurify'
import {useI18n} from 'vue-i18n'

import {useProjectStore} from '@/stores/projects'
import {useProjectWikiPageService} from '@/services/projectWikiPage'
import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import {PERMISSIONS} from '@/constants/permissions'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import Card from '@/components/misc/Card.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps<{
	projectId: number,
	isLoadingProject: boolean,
	viewId: number,
}>()

const {t} = useI18n()
const projectStore = useProjectStore()
const wikiService = useProjectWikiPageService()

const project = computed(() => projectStore.projects[props.projectId])
const canWrite = computed(() => (project.value?.maxPermission ?? 0) >= PERMISSIONS.READ_WRITE)

const pages = ref<IProjectWikiPage[]>([])
const homePage = computed(() => {
	if (pages.value.length === 0) return null
	return pages.value.find(p => p.isHome) || pages.value[0]
})

async function loadWikiPages() {
	if (!props.projectId || props.projectId <= 0) return
	try {
		const res = await wikiService.getAll(props.projectId)
		pages.value = res.items
	} catch (e) {
		console.error('Failed to load wiki pages for overview:', e)
	}
}

watch(() => props.projectId, loadWikiPages)
onMounted(loadWikiPages)

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

const htmlWikiContent = computed(() => {
	const c = homePage.value?.content || ''
	if (!c) return ''
	return DOMPurify.sanitize(c, {ADD_ATTR: ['target']})
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
}
</style>
