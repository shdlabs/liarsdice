import React from 'react'
import { SidebarDetailsProps } from '../types/props.d'
import Players from './players'

// SideBarDetails component
function SidebarDetails(props: SidebarDetailsProps) {
  // Extracts props.
  const { round, ante, pot } = props

  // Renders this markup
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'start',
        flexDirection: 'column',
        backgroundColor: 'var(--modals)',
        width: '100%',
        flexGrow: '1',
      }}
    >
      <div
        className="details"
        style={{
          padding: '16px 10px',
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
      <Players />
    </div>
  )
}
export default SidebarDetails
