import type {IAbstract} from './IAbstract'
import type {IFile} from './IFile'
import type {IUser} from './IUser'

export interface IProjectWikiPageAttachment extends IAbstract {
	id: number
	pageId: number
	createdBy?: IUser
	file: IFile
	created: Date
}
