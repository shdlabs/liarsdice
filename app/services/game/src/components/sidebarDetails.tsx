import React from 'react'

interface SidebarDetailsProps {
  round: number
  ante?: number
  pot?: number
}
const SidebarDetails = (props: SidebarDetailsProps) => {
  const { round, ante, pot } = props
  return (
    <div
      className="details"
      style={{
        display: 'flex',
        alignItems: 'start',
        flexDirection: 'column',
        backgroundColor: 'var(--modals)',
        borderRadius: '0 8px 8px 0',
        margin: '42px 0 42px 0',
        padding: '12px',
        width: '80%',
      }}
    >
      <div className="d-flex">
        <strong className="details__title mr-6">Round:</strong>
        {round ? round : '-'}
      </div>
      <div className="d-flex">
        {ante ? (
          <>
            <strong className="details__title mr-6">Ante:</strong>
            {ante} U$D
          </>
        ) : (
          ''
        )}
      </div>
      <div className="d-flex">
        {pot ? (
          <>
            <strong className="details__title mr-6">Pot:</strong>
            {pot} U$D
          </>
        ) : (
          ''
        )}
      </div>
    </div>
  )
}
export default SidebarDetails
