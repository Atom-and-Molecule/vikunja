import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import type {IUser} from '@/modelTypes/IUser'

export default class ProjectWikiPage implements IProjectWikiPage {
	id = 0
	projectId = 0
	parentPageId = 0
	title = ''
	content = ''
	position = 0
	isHome = false

	createdBy?: IUser
	updatedBy?: IUser

	created = new Date()
	updated = new Date()
}
