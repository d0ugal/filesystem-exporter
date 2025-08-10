# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in filesystem-exporter, please follow these steps:

### 1. **DO NOT** create a public GitHub issue
Security vulnerabilities should be reported privately to avoid potential exploitation.

### 2. Report the vulnerability
Please email security details to: [security@example.com](mailto:security@example.com)

Include the following information:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)
- Your contact information

### 3. Response timeline
- **Initial response**: Within 48 hours
- **Status update**: Within 1 week
- **Resolution**: As quickly as possible, typically within 30 days

### 4. Disclosure
- Security issues will be disclosed via GitHub Security Advisories
- CVE numbers will be requested when appropriate
- Patches will be released as soon as possible

## Security Best Practices

### For Users
- Keep filesystem-exporter updated to the latest version
- Review configuration files for sensitive information
- Use appropriate file permissions for configuration files
- Monitor logs for unusual activity
- Run the container with minimal required privileges

### For Contributors
- Follow secure coding practices
- Validate all user inputs
- Use parameterized queries and avoid command injection
- Keep dependencies updated
- Review code for potential security issues

## Security Features

filesystem-exporter includes several security features:

- **Input validation**: All configuration inputs are validated
- **Path sanitization**: Mount points and directory paths are sanitized
- **Read-only access**: Container runs with read-only filesystem access where possible
- **Minimal attack surface**: Small, focused binary with minimal dependencies
- **Structured logging**: Secure logging without sensitive information exposure

## Dependencies

We regularly update dependencies to address security vulnerabilities:

- Automated dependency scanning in CI/CD
- Regular security audits
- Prompt updates for critical vulnerabilities

## Responsible Disclosure

We appreciate security researchers who:
- Report vulnerabilities privately
- Allow reasonable time for fixes
- Work with us to coordinate disclosure
- Follow responsible disclosure practices

Thank you for helping keep filesystem-exporter secure!

