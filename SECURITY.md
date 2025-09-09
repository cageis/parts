# Security Policy

## Supported Versions

We actively maintain and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Parts seriously. If you believe you have found a security vulnerability, please report it to us as described below.

**Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

1. **Email:** Send details to [security@yourproject.com] (replace with actual email)
2. **GitHub Security Advisory:** Use GitHub's private security reporting feature
3. **Include the following information:**
   - Type of issue (buffer overflow, SQL injection, cross-site scripting, etc.)
   - Full paths of source file(s) related to the issue
   - Location of the affected source code (tag/branch/commit or direct URL)
   - Any special configuration required to reproduce the issue
   - Step-by-step instructions to reproduce the issue
   - Proof-of-concept or exploit code (if possible)
   - Impact of the issue, including how an attacker might exploit it

### Response Timeline

- **Initial Response:** Within 48 hours of receiving the report
- **Assessment:** Within 1 week, we'll provide a detailed response indicating next steps
- **Resolution:** Security issues will be prioritized and fixed as quickly as possible
- **Disclosure:** We'll coordinate disclosure timing with you

### Security Measures

Parts implements several security measures:

1. **Input Validation:** All file paths and user inputs are validated
2. **Path Traversal Protection:** File paths are checked to prevent directory traversal
3. **Safe File Operations:** All file operations use safe APIs
4. **Dependency Scanning:** Dependencies are regularly scanned for vulnerabilities
5. **Code Analysis:** Static analysis tools are used to detect potential issues

### Security Best Practices for Users

When using Parts:

1. **Validate Inputs:** Ensure file paths and directories are trusted
2. **Use Dry-Run:** Test changes with `--dry-run` before applying
3. **Backup Files:** Keep backups of important configuration files
4. **Check Permissions:** Ensure appropriate file permissions on config files
5. **Review Changes:** Always review generated content before using it

### Scope

This security policy applies to:
- The Parts CLI tool
- All official releases and pre-releases
- The main repository and official documentation

### Out of Scope

The following are generally out of scope:
- Issues in third-party dependencies (report to the respective projects)
- Issues requiring physical access to the machine
- Issues in user-created configuration files or scripts
- Social engineering attacks

### Recognition

We appreciate the security research community and will acknowledge researchers who responsibly disclose vulnerabilities to us. With your permission, we'll:

- Thank you in the security advisory
- Include your name in our security acknowledgments
- Provide attribution in release notes (if desired)

### Questions

If you have questions about this security policy, please create a GitHub issue or contact the maintainers.