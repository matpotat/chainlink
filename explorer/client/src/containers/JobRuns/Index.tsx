import React, { useState, useEffect } from 'react'
import { connect } from 'react-redux'
import { bindActionCreators, Dispatch } from 'redux'
import build from 'redux-object'
import {
  createStyles,
  Theme,
  withStyles,
  WithStyles
} from '@material-ui/core/styles'
import List from '../../components/JobRuns/List'
import { getJobRuns } from '../../actions/jobRuns'
import { IState } from '../../reducers'
import { Query } from '../../reducers/search'
import { ChangePageEvent } from '../../components/Table'

const EMPTY_MSG =
  "We couldn't find any results for your search query. Try again with the job id, run id, requester, requester id or transaction hash"

const styles = ({ spacing, breakpoints }: Theme) =>
  createStyles({
    container: {
      paddingTop: spacing.unit * 2,
      paddingBottom: spacing.unit * 2,
      paddingLeft: spacing.unit * 2,
      paddingRight: spacing.unit * 2,
      [breakpoints.up('sm')]: {
        paddingTop: spacing.unit * 3,
        paddingBottom: spacing.unit * 3,
        paddingLeft: spacing.unit * 3,
        paddingRight: spacing.unit * 3
      }
    }
  })

interface IOwnProps {
  path: string
}

interface IStateProps {
  rowsPerPage: number
  query?: string
  jobRuns?: IJobRun[]
  count?: number
}

interface IDispatchProps {
  getJobRuns: (query: Query, page: number, size: number) => void
}

interface IProps
  extends WithStyles<typeof styles>,
    IOwnProps,
    IStateProps,
    IDispatchProps {}

const Index = withStyles(styles)(
  ({ getJobRuns, query, rowsPerPage, classes, jobRuns, count }: IProps) => {
    const [currentPage, setCurrentPage] = useState(0)
    const onChangePage = (_event: ChangePageEvent, page: number) => {
      setCurrentPage(page)
      getJobRuns(query, page + 1, rowsPerPage)
    }

    useEffect(() => {
      getJobRuns(query, currentPage + 1, rowsPerPage)
    }, [getJobRuns, query, currentPage, rowsPerPage])

    return (
      <div className={classes.container}>
        <List
          currentPage={currentPage}
          jobRuns={jobRuns}
          count={count}
          onChangePage={onChangePage}
          emptyMsg={EMPTY_MSG}
        />
      </div>
    )
  }
)

const jobRunsSelector = ({
  jobRunsIndex,
  jobRuns,
  chainlinkNodes
}: IState): IJobRun[] | undefined => {
  if (jobRunsIndex.items) {
    return jobRunsIndex.items.map((id: string) => {
      const document = {
        jobRuns: jobRuns.items,
        chainlinkNodes: chainlinkNodes.items
      }
      return build(document, 'jobRuns', id)
    })
  }
}

const jobRunsCountSelector = (state: IState) => {
  return state.jobRunsIndex.count
}

const mapStateToProps = (state: IState): IStateProps => {
  return {
    rowsPerPage: 10,
    query: state.search.query,
    jobRuns: jobRunsSelector(state),
    count: jobRunsCountSelector(state)
  }
}

const mapDispatchToProps = (dispatch: Dispatch): IDispatchProps =>
  bindActionCreators({ getJobRuns }, dispatch)

const ConnectedIndex = connect<IStateProps, IDispatchProps, IOwnProps, IState>(
  mapStateToProps,
  mapDispatchToProps
)(Index)

export default ConnectedIndex
