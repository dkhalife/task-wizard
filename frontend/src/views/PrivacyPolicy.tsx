import {
  Container,
  Divider,
  List,
  ListItem,
  ListItemDecorator,
  Typography,
} from '@mui/joy'
import {
  CheckCircle,
  RemoveCircle,
} from '@mui/icons-material'
import React from 'react'

export class PrivacyPolicy extends React.Component {
  render(): React.ReactNode {
    return (
      <Container
        maxWidth='md'
        sx={{ py: 4, px: 2 }}
      >
        <Typography level='h1'>Privacy Policy</Typography>
        <Typography
          level='body-sm'
          sx={{ mt: 1, mb: 3, color: 'text.tertiary' }}
        >
          Last updated: March 2026
        </Typography>

        <Section title='Overview'>
          <Typography>
            Task Wizard is a self-hostable, privacy-focused task management
            application. Because you can host it on your own infrastructure, you
            retain full control over your data. No user information is sent to the
            Task Wizard maintainers or any centralized service operated by them.
          </Typography>
        </Section>

        <Section title='Data We Collect'>
          <Typography sx={{ mb: 1 }}>
            Task Wizard is designed to collect the minimum data necessary to
            function:
          </Typography>
          <BulletList
            icon={<CheckCircle color='success' />}
            items={[
              'Authentication identifiers from Microsoft Entra ID (directory ID and object ID) — used solely to identify your account',
              'Task data you create: titles, due dates, recurrence rules, completion status, and labels',
              'Notification preferences and scheduled notification metadata',
              'Request metadata for operational logging: IP address, user agent, HTTP method, route, and status code (request bodies and tokens are never logged)',
            ]}
          />
        </Section>

        <Section title='Data We Do NOT Collect'>
          <BulletList
            icon={<RemoveCircle color='disabled' />}
            items={[
              'Personal names, email addresses, or passwords — these fields were explicitly omitted from the data model',
              'Tracking cookies, analytics identifiers, or advertising data',
            ]}
          />
        </Section>

        <Section title='Optional Telemetry'>
          <Typography>
            Task Wizard includes optional, opt-in telemetry powered by Azure
            Application Insights. Telemetry is used solely to monitor application
            health, track errors, and improve the user experience.
          </Typography>
          <Typography sx={{ mt: 1, fontWeight: 600 }}>
            On the Android app, telemetry is disabled by default and must be
            explicitly enabled by the user in Settings → Analytics.
          </Typography>
          <Typography sx={{ mt: 1 }}>
            When telemetry is disabled, the Android app sends a
            Do-Not-Track (DNT) header with every API request. The backend
            respects this header and skips request-level telemetry for that user.
          </Typography>
          <Typography sx={{ mt: 1 }}>
            When enabled, the following anonymous data may be collected:
          </Typography>
          <BulletList
            icon={<CheckCircle color='success' />}
            items={[
              'Application errors, warnings, and crash reports',
              'HTTP request metadata (method, route, status code, duration)',
              'Build version, build number, and commit hash for the running software',
              'A randomly generated device identifier (not linked to any account or personal information)',
            ]}
          />
          <Typography sx={{ mt: 1 }}>
            An additional &ldquo;Debug logging&rdquo; toggle (visible only when
            telemetry is enabled) sends more detailed diagnostic data to help
            troubleshoot specific issues.
          </Typography>
          <Typography sx={{ mt: 1 }}>
            No task content, user identifiers, authentication tokens, or
            personally identifiable information is ever included in telemetry
            data.
          </Typography>
        </Section>

        <Section title='How Your Data Is Stored'>
          <Typography>
            All data is stored in a database (SQLite by default, or MySQL) on
            a backend server.
          </Typography>
          <Typography sx={{ mt: 1 }}>
            The database is not encrypted at rest by default. Server
            administrators are encouraged to apply operating-system-level or
            disk-level encryption to protect stored data.
          </Typography>
        </Section>

        <Section title='Authentication and Security'>
          <BulletList
            icon={<CheckCircle color='success' />}
            items={[
              'Authentication is delegated to Microsoft Entra ID using OAuth 2.0 and OpenID Connect — no passwords are stored or processed by the application',
              'JWT tokens are verified against Entra ID\'s public signing keys (JWKS) on every request',
              'WebSocket connections are authenticated using the same token verification',
              'Rate limiting is applied (300 requests per minute per IP address) to mitigate abuse',
              'All database queries use parameterized statements to prevent SQL injection',
              'CORS is configurable per deployment to restrict cross-origin access',
              'HTTPS is recommended and should be configured via a reverse proxy in front of the application',
            ]}
          />
        </Section>

        <Section title='Third-Party Services'>
          <BulletList
            icon={<CheckCircle color='success' />}
            items={[
              'Microsoft Entra ID - contacted for authentication only; no task or personal data is shared',
              'Azure Application Insights (optional) - when telemetry is enabled, anonymous operational data (errors, request metadata, build info) is sent to the configured Application Insights instance; no task content or personal data is included',
              'Gotify or webhook endpoints (optional) - if configured by you, only minimal task completion text is sent to the endpoint you choose',
            ]}
          />
          <Typography sx={{ mt: 1 }}>
            No other external services are contacted by the application.
          </Typography>
        </Section>

        <Section title='Data Retention and Deletion'>
          <BulletList
            icon={<CheckCircle color='success' />}
            items={[
              'Sent notifications are automatically deleted within 10 minutes',
              'Deleting your account removes all associated data including tasks, labels, and notifications',
              'As a self-hostable application, the server administrator has full control over data retention, backups, and purging',
            ]}
          />
        </Section>

        <Section title='Open Source and Transparency'>
          <Typography>
            Task Wizard is open-source software. The entire codebase is publicly
            available and can be audited by anyone. Automated security scanning
            is performed via CodeQL and dependency updates are managed through
            Dependabot.
          </Typography>
        </Section>

        <Divider sx={{ my: 3 }} />

        <Section title='Disclaimers'>
          <Typography>
            This software is provided &ldquo;as is&rdquo;, without warranty of
            any kind, express or implied. While the project follows security best
            practices, no system is perfect and vulnerabilities may exist.
          </Typography>
          <Typography sx={{ mt: 1, fontWeight: 600 }}>
            Use at your own risk.
          </Typography>
          <Typography sx={{ mt: 1 }}>
            The maintainers are not liable for any data loss, security breaches,
            or damages arising from the use of this software. Because Task
            Wizard is self-hostable, the security of your deployment ultimately
            depends on you: keep your server, reverse proxy, operating system,
            and dependencies up to date.
          </Typography>
        </Section>

        <Section title='Changes to This Policy'>
          <Typography>
            This policy may be updated as the application evolves. Changes will
            be reflected in the &ldquo;last updated&rdquo; date at the top of
            this page.
          </Typography>
        </Section>
      </Container>
    )
  }
}

interface SectionProps {
  title: string
  children: React.ReactNode
}

class Section extends React.Component<SectionProps> {
  render(): React.ReactNode {
    return (
      <section style={{ marginBottom: 24 }}>
        <Typography
          level='h3'
          sx={{ mb: 1 }}
        >
          {this.props.title}
        </Typography>
        {this.props.children}
      </section>
    )
  }
}

interface BulletListProps {
  icon: React.ReactNode
  items: string[]
}

class BulletList extends React.Component<BulletListProps> {
  render(): React.ReactNode {
    return (
      <List size='sm'>
        {this.props.items.map((item, index) => (
          <ListItem key={index}>
            <ListItemDecorator sx={{ alignSelf: 'flex-start', mt: 0.5 }}>
              {this.props.icon}
            </ListItemDecorator>
            {item}
          </ListItem>
        ))}
      </List>
    )
  }
}
