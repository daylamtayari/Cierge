import { Link } from 'react-router-dom'
import Layout from '../components/Layout'

export default function NotFound() {
  return (
    <Layout>
      <div className="container">
        <div className="notfound">
          <p className="notfound-eyebrow">404</p>
          <h1 className="notfound-title">Page not found</h1>
          <p className="notfound-body">
            We couldn't find the page you were looking for. The link may be
            incorrect, or the page may have moved.
          </p>
          <Link to="/" className="btn btn-primary">Back to bookings</Link>
        </div>
      </div>
    </Layout>
  )
}
