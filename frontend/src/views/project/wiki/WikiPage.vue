<!-- eslint-disable vue/no-v-html -->
<template>
	<div class="wiki-view p-4">
		<div class="columns is-variable is-4">
			<!-- Sidebar: Page Tree -->
			<div class="column is-3">
				<WikiTree
					:pages="pages"
					:current-page-id="currentPage?.id"
					:can-write="canWrite"
					@selectPage="handleSelectPage"
					@newPage="handleNewPage"
				/>
			</div>

			<!-- Main Area: Page View / Edit -->
			<div class="column is-9">
				<Card
					v-if="currentPage && !isEditing"
					:has-content="false"
					class="wiki-page-card"
				>
					<header class="wiki-page-header is-flex is-justify-content-between is-align-items-center pbe-3 mbe-4 border-bottom">
						<div class="is-flex is-align-items-center gap-2">
							<BaseButton
								class="is-small is-light"
								:to="{ name: 'project.index', params: { projectId } }"
							>
								<span class="icon is-small">
									<Icon icon="arrow-left" />
								</span>
								<span>{{ $t('project.overview.title') }}</span>
							</BaseButton>

							<span
								v-if="currentPage.isHome"
								class="tag is-info is-light"
							>
								{{ $t('project.wiki.homePage') }}
							</span>
						</div>

						<div class="is-flex is-align-items-center gap-2">
							<BaseButton
								class="is-small is-light"
								@click="handleOpenHistory"
							>
								<span class="icon is-small">
									<Icon icon="history" />
								</span>
								<span>{{ $t('project.wiki.history') }}</span>
							</BaseButton>

							<template v-if="canWrite">
								<BaseButton
									v-if="!currentPage.isHome"
									class="is-small is-light"
									@click="handleSetAsHome"
								>
									{{ $t('project.wiki.setAsHome') }}
								</BaseButton>

								<BaseButton
									class="is-small is-primary"
									@click="isEditing = true"
								>
									<span class="icon is-small">
										<Icon icon="pen" />
									</span>
									<span>{{ $t('project.wiki.editPage') }}</span>
								</BaseButton>

								<BaseButton
									class="is-small is-danger is-light"
									@click="handleDeletePage"
								>
									<span class="icon is-small">
										<Icon icon="trash" />
									</span>
								</BaseButton>
							</template>
						</div>
					</header>

					<h1 class="title is-3 mbe-2">
						{{ currentPage.title }}
					</h1>

					<p
						v-if="currentPage.updated"
						class="is-size-7 has-text-grey mbe-4"
					>
						{{ $t('project.wiki.lastUpdated', {
							user: currentPage.updatedBy?.username || currentPage.createdBy?.username || '',
							date: formatDisplayDate(currentPage.updated),
						}) }}
					</p>

					<div
						v-if="htmlContent !== ''"
						class="content wiki-content"
						v-html="htmlContent"
					/>
					<p
						v-else
						class="is-italic has-text-grey"
					>
						{{ $t('project.overview.wikiComingSoon') }}
					</p>

					<!-- Attachments Section -->
					<div
						v-if="currentPage.id > 0"
						class="wiki-attachments-section mbs-5 pbs-4 border-top"
					>
						<div class="is-flex is-justify-content-between is-align-items-center mbe-3">
							<h3 class="title is-6 mbe-0 is-flex is-align-items-center gap-2">
								<span class="icon is-small">
									<Icon icon="paperclip" />
								</span>
								<span>{{ $t('project.wiki.attachments') }} ({{ attachments.length }})</span>
							</h3>
							<div
								v-if="canWrite"
								class="is-flex is-align-items-center"
							>
								<input
									ref="fileInputRef"
									type="file"
									multiple
									class="is-hidden"
									@change="handleFileUpload"
								>
								<BaseButton
									class="is-small is-light"
									:disabled="isUploading"
									@click="fileInputRef?.click()"
								>
									<span class="icon is-small">
										<Icon icon="plus" />
									</span>
									<span>{{ $t('project.wiki.uploadAttachment') }}</span>
								</BaseButton>
							</div>
						</div>

						<div
							v-if="attachments.length === 0"
							class="is-size-7 is-italic has-text-grey"
						>
							{{ $t('project.wiki.noAttachments') }}
						</div>
						<div
							v-else
							class="attachments-list"
						>
							<div
								v-for="att in attachments"
								:key="att.id"
								class="attachment-item is-flex is-justify-content-between is-align-items-center p-2 mb-2"
							>
								<div class="is-flex is-align-items-center gap-2">
									<span class="icon has-text-grey">
										<Icon icon="file" />
									</span>
									<div>
										<span class="has-text-weight-semibold is-size-7">{{ att.file.name }}</span>
										<span class="is-size-7 has-text-grey mis-2">({{ getHumanSize(att.file.size) }})</span>
									</div>
								</div>
								<div class="buttons is-right mbe-0">
									<BaseButton
										class="is-small is-light"
										@click="handleDownloadAttachment(att)"
									>
										<span class="icon is-small">
											<Icon icon="download" />
										</span>
									</BaseButton>
									<BaseButton
										v-if="canWrite"
										class="is-small is-danger is-light"
										@click="handleDeleteAttachment(att)"
									>
										<span class="icon is-small">
											<Icon icon="trash" />
										</span>
									</BaseButton>
								</div>
							</div>
						</div>
					</div>
				</Card>

				<!-- Editing Form -->
				<Card
					v-else-if="currentPage && isEditing"
					:has-content="false"
					class="wiki-page-edit-card"
				>
					<h2 class="title is-4 mbe-4">
						{{ currentPage.id === 0 ? $t('project.wiki.newPage') : $t('project.wiki.editPage') }}
					</h2>

					<FormField
						id="wiki-title"
						v-model="editTitle"
						:label="$t('project.wiki.pageTitle')"
						:placeholder="$t('project.wiki.titlePlaceholder')"
						class="mbe-4"
					/>

					<FormField
						:label="$t('project.wiki.title')"
						class="mbe-4"
					>
						<Editor
							id="wiki-content-editor"
							v-model="editContent"
							:placeholder="$t('project.wiki.contentPlaceholder')"
						/>
					</FormField>

					<div class="field is-grouped is-justify-content-flex-end">
						<div class="control">
							<BaseButton
								class="button is-light"
								@click="handleCancelEdit"
							>
								{{ $t('project.wiki.cancel') }}
							</BaseButton>
						</div>
						<div class="control">
							<BaseButton
								class="button is-primary"
								:class="{'is-loading': isSaving}"
								:disabled="!editTitle.trim() || isSaving"
								@click="handleSavePage"
							>
								{{ $t('project.wiki.savePage') }}
							</BaseButton>
						</div>
					</div>
				</Card>

				<!-- Empty State when no page exists -->
				<Card
					v-else
					class="has-text-centered p-6"
				>
					<span class="icon is-large has-text-grey-light mbe-3">
						<Icon
							icon="file"
							size="3x"
						/>
					</span>
					<p class="is-size-5 has-text-weight-semibold mbe-2">
						{{ $t('project.wiki.noPages') }}
					</p>
					<div
						v-if="canWrite"
						class="mbe-4"
					>
						<BaseButton
							class="button is-primary"
							@click="handleNewPage()"
						>
							<span class="icon">
								<Icon icon="plus" />
							</span>
							<span>{{ $t('project.wiki.createFirstPage') }}</span>
						</BaseButton>
					</div>
				</Card>
			</div>
		</div>

		<!-- Revision History Modal -->
		<Modal
			:enabled="showHistoryModal"
			wide
			@close="showHistoryModal = false"
		>
			<Card
				class="has-no-shadow"
				:title="$t('project.wiki.revisionHistory')"
				:show-close="true"
				:has-content="false"
				@close="showHistoryModal = false"
			>
				<div
					v-if="loadingRevisions"
					class="has-text-centered p-6"
				>
					<span class="is-italic has-text-grey">{{ $t('misc.loading') }}</span>
				</div>
				<div
					v-else-if="revisions.length === 0"
					class="has-text-centered p-6"
				>
					<p class="is-italic has-text-grey">
						{{ $t('project.wiki.noRevisions') }}
					</p>
				</div>
				<div
					v-else
					class="columns is-variable is-3 p-4"
				>
					<!-- Revision list -->
					<div class="column is-5">
						<div class="menu">
							<ul class="menu-list">
								<li
									v-for="rev in revisions"
									:key="rev.id"
								>
									<a
										:class="{ 'is-active': selectedRevision?.id === rev.id }"
										class="is-flex is-flex-direction-column py-2"
										@click="handleSelectRevision(rev)"
									>
										<span class="has-text-weight-semibold">{{ rev.title }}</span>
										<span class="is-size-7 has-text-grey">
											{{ formatDisplayDate(rev.created) }}
											<template v-if="rev.createdBy">
												&bull; {{ rev.createdBy.name || rev.createdBy.username }}
											</template>
										</span>
									</a>
								</li>
							</ul>
						</div>
					</div>

					<!-- Selected revision preview -->
					<div class="column is-7">
						<div
							v-if="selectedRevision"
							class="box p-4"
						>
							<div class="is-flex is-justify-content-between is-align-items-center mbe-3">
								<div>
									<h3 class="title is-5 mbe-1">
										{{ selectedRevision.title }}
									</h3>
									<p class="is-size-7 has-text-grey">
										{{ formatDisplayDate(selectedRevision.created) }}
										<template v-if="selectedRevision.createdBy">
											&bull; {{ selectedRevision.createdBy.name || selectedRevision.createdBy.username }}
										</template>
									</p>
								</div>
								<BaseButton
									v-if="canWrite"
									class="is-small is-warning"
									@click="handleRestoreRevision(selectedRevision)"
								>
									<span class="icon is-small">
										<Icon icon="history" />
									</span>
									<span>{{ $t('project.wiki.restore') }}</span>
								</BaseButton>
							</div>
							<hr class="my-2">
							<div
								class="content wiki-content is-size-7"
								v-html="revisionPreviewHtml"
							/>
						</div>
					</div>
				</div>

				<template #footer>
					<BaseButton
						class="button is-light"
						@click="showHistoryModal = false"
					>
						{{ $t('misc.close') }}
					</BaseButton>
				</template>
			</Card>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import DOMPurify from 'dompurify'

import {useProjectStore} from '@/stores/projects'
import {useProjectWikiPageService} from '@/services/projectWikiPage'
import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import type {IProjectWikiPageRevision} from '@/modelTypes/IProjectWikiPageRevision'
import type {IProjectWikiPageAttachment} from '@/modelTypes/IProjectWikiPageAttachment'
import ProjectWikiPage from '@/models/projectWikiPage'
import {PERMISSIONS} from '@/constants/permissions'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {getHumanSize} from '@/helpers/getHumanSize'
import {error, success} from '@/message'

import Card from '@/components/misc/Card.vue'
import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import FormField from '@/components/input/FormField.vue'
import Editor from '@/components/input/AsyncEditor'
import WikiTree from '@/components/project/wiki/WikiTree.vue'

const props = defineProps<{
	projectId: number,
	pageId: number,
}>()

const {t} = useI18n()
const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const wikiService = useProjectWikiPageService()

const project = computed(() => projectStore.projects[props.projectId])
const canWrite = computed(() => (project.value?.maxPermission ?? 0) >= PERMISSIONS.READ_WRITE)

const pages = ref<IProjectWikiPage[]>([])
const currentPage = ref<IProjectWikiPage | null>(null)
const isEditing = ref(false)
const isSaving = ref(false)

const editTitle = ref('')
const editContent = ref('')
const editParentId = ref(0)

const htmlContent = computed(() => {
	const c = currentPage.value?.content || ''
	if (!c) return ''
	return DOMPurify.sanitize(c, {ADD_ATTR: ['target']})
})

const showHistoryModal = ref(false)
const revisions = ref<IProjectWikiPageRevision[]>([])
const loadingRevisions = ref(false)
const selectedRevision = ref<IProjectWikiPageRevision | null>(null)

const revisionPreviewHtml = computed(() => {
	const c = selectedRevision.value?.content || ''
	if (!c) return ''
	return DOMPurify.sanitize(c, {ADD_ATTR: ['target']})
})

const attachments = ref<IProjectWikiPageAttachment[]>([])
const fileInputRef = ref<HTMLInputElement | null>(null)
const isUploading = ref(false)

async function loadPages() {
	try {
		const res = await wikiService.getAll(props.projectId)
		pages.value = res.items
		resolveActivePage()
	} catch (e) {
		console.error('Failed to load wiki pages:', e)
	}
}

function resolveActivePage() {
	if (props.pageId > 0) {
		const found = pages.value.find(p => p.id === props.pageId)
		if (found) {
			currentPage.value = found
			editTitle.value = found.title
			editContent.value = found.content
			editParentId.value = found.parentPageId
			return
		}
	}

	// Route without pageId: find home page or first page
	const home = pages.value.find(p => p.isHome) || pages.value[0]
	if (home) {
		currentPage.value = home
		editTitle.value = home.title
		editContent.value = home.content
		editParentId.value = home.parentPageId
	} else {
		currentPage.value = null
	}
}

watch(() => props.pageId, resolveActivePage)

function handleSelectPage(pageId: number) {
	isEditing.value = false
	router.push({
		name: 'project.wiki.page',
		params: {
			projectId: props.projectId,
			pageId,
		},
	})
}

function handleNewPage(parentPageId?: number) {
	const newPage = new ProjectWikiPage()
	newPage.projectId = props.projectId
	newPage.parentPageId = parentPageId ?? 0
	currentPage.value = newPage
	editTitle.value = ''
	editContent.value = ''
	editParentId.value = parentPageId ?? 0
	isEditing.value = true
}

function handleCancelEdit() {
	isEditing.value = false
	resolveActivePage()
}

async function handleSavePage() {
	if (!currentPage.value || !editTitle.value.trim()) return

	isSaving.value = true
	try {
		if (currentPage.value.id === 0) {
			// Create
			const created = await wikiService.create(props.projectId, {
				title: editTitle.value.trim(),
				content: editContent.value,
				parentPageId: editParentId.value,
			})
			await loadPages()
			isEditing.value = false
			success({message: t('project.wiki.savedSuccessfully')})
			router.push({
				name: 'project.wiki.page',
				params: {
					projectId: props.projectId,
					pageId: created.id,
				},
			})
		} else {
			// Update
			currentPage.value.title = editTitle.value.trim()
			currentPage.value.content = editContent.value
			await wikiService.update(props.projectId, currentPage.value)
			await loadPages()
			isEditing.value = false
			success({message: t('project.wiki.savedSuccessfully')})
		}
	} catch (e) {
		console.error('Failed to save wiki page:', e)
		error(e)
	} finally {
		isSaving.value = false
	}
}

async function handleSetAsHome() {
	if (!currentPage.value || currentPage.value.id === 0) return
	try {
		currentPage.value.isHome = true
		await wikiService.update(props.projectId, currentPage.value)
		await loadPages()
	} catch (e) {
		console.error('Failed to set home page:', e)
	}
}

async function handleOpenHistory() {
	if (!currentPage.value || currentPage.value.id === 0) return
	showHistoryModal.value = true
	loadingRevisions.value = true
	selectedRevision.value = null
	try {
		const res = await wikiService.getRevisions(props.projectId, currentPage.value.id)
		revisions.value = res.items
		if (revisions.value.length > 0) {
			selectedRevision.value = revisions.value[0]
		}
	} catch (e) {
		console.error('Failed to load wiki page revisions:', e)
	} finally {
		loadingRevisions.value = false
	}
}

function handleSelectRevision(rev: IProjectWikiPageRevision) {
	selectedRevision.value = rev
}

function handleRestoreRevision(rev: IProjectWikiPageRevision) {
	if (!confirm(t('project.wiki.restoreConfirm'))) return
	editTitle.value = rev.title
	editContent.value = rev.content
	showHistoryModal.value = false
	isEditing.value = true
}

async function handleDeletePage() {
	if (!currentPage.value || currentPage.value.id === 0) return
	if (!confirm(t('project.wiki.deleteConfirm'))) return

	try {
		await wikiService.remove(props.projectId, currentPage.value.id)
		await loadPages()
		router.push({
			name: 'project.wiki',
			params: {projectId: props.projectId},
		})
	} catch (e) {
		console.error('Failed to delete wiki page:', e)
	}
}

async function loadAttachments() {
	if (!currentPage.value || currentPage.value.id === 0) {
		attachments.value = []
		return
	}
	try {
		const res = await wikiService.getAttachments(props.projectId, currentPage.value.id)
		attachments.value = res.items
	} catch (e) {
		console.error('Failed to load wiki page attachments:', e)
	}
}

watch(() => currentPage.value?.id, () => {
	loadAttachments()
})

async function handleFileUpload(event: Event) {
	const target = event.target as HTMLInputElement
	if (!target.files || target.files.length === 0 || !currentPage.value || currentPage.value.id === 0) return

	isUploading.value = true
	try {
		const filesToUpload = Array.from(target.files)
		await wikiService.uploadAttachments(props.projectId, currentPage.value.id, filesToUpload)
		await loadAttachments()
		target.value = ''
	} catch (e) {
		console.error('Failed to upload attachments:', e)
	} finally {
		isUploading.value = false
	}
}

async function handleDownloadAttachment(att: IProjectWikiPageAttachment) {
	if (!currentPage.value || currentPage.value.id === 0) return
	try {
		await wikiService.downloadAttachment(props.projectId, currentPage.value.id, att)
	} catch (e) {
		console.error('Failed to download attachment:', e)
	}
}

async function handleDeleteAttachment(att: IProjectWikiPageAttachment) {
	if (!currentPage.value || currentPage.value.id === 0) return
	if (!confirm(t('project.wiki.deleteAttachmentConfirm', {filename: att.file.name}))) return
	try {
		await wikiService.deleteAttachment(props.projectId, currentPage.value.id, att.id)
		await loadAttachments()
	} catch (e) {
		console.error('Failed to delete attachment:', e)
	}
}

onMounted(() => {
	loadPages()
	if (route.query.edit === 'true' && canWrite.value) {
		isEditing.value = true
	} else if (route.query.new === 'true' && canWrite.value) {
		handleNewPage()
	}
})
</script>

<style lang="scss" scoped>
.wiki-view {
	max-inline-size: 1400px;
	margin-inline: auto;
}

.wiki-content {
	line-height: 1.6;
}

.border-top {
	border-block-start: 1px solid var(--border);
}

.border-bottom {
	border-block-end: 1px solid var(--border);
}

.attachments-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.attachment-item {
	background-color: var(--white);
	border: 1px solid var(--border);
	border-radius: $radius;
}
</style>
