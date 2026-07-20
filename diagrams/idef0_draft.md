# IDEF0 Draft for EngProject

## System

English learning web platform with Moodle integration.

## A-0

**Function:** Manage the English learning process in the web platform

**Inputs:**
- user requests for login and learning
- educational content for courses, decks, cards, and texts
- personal vocabulary entries and translations
- Moodle account data

**Controls:**
- learning methodology and course structure
- access rules and user roles
- training logic and progress rules
- platform configuration
- integration rules for Moodle

**Outputs:**
- accessible learning courses
- completed training sessions
- user progress data
- personal vocabulary and saved translations
- reports on learning results

**Mechanisms:**
- student
- administrator
- frontend application
- backend API
- PostgreSQL database
- Moodle integration
- dictionary/translation services

## A0 Decomposition

### A1. Authorize user and provide access

**Inputs:**
- login request
- Moodle credentials
- user session request

**Controls:**
- authentication rules
- authorization rules
- Moodle-only auth mode

**Outputs:**
- authenticated user session
- access rights to courses
- user profile data

**Mechanisms:**
- frontend login page
- auth handlers
- JWT logic
- Moodle service
- user database

### A2. Manage learning content

**Inputs:**
- course data
- deck data
- card data
- uploaded files

**Controls:**
- content structure rules
- admin permissions
- import validation rules

**Outputs:**
- published courses
- updated decks and cards
- imported learning materials

**Mechanisms:**
- administrator
- course/deck/card handlers
- upload handlers
- content services
- database

### A3. Conduct learning and training

**Inputs:**
- selected course or deck
- selected cards and texts
- user answers in training

**Controls:**
- training scenarios
- sequencing rules
- session flow logic

**Outputs:**
- training results
- completed sessions
- updated card learning state

**Mechanisms:**
- student
- frontend learning interface
- training session handlers
- training flow services
- database

### A4. Maintain personal vocabulary

**Inputs:**
- words added by user
- personal translations
- dictionary lookup requests

**Controls:**
- vocabulary management rules
- duplicate handling rules
- translation saving rules

**Outputs:**
- personal vocabulary list
- saved translations
- enriched language reference data

**Mechanisms:**
- student
- personal vocabulary handlers
- translation handlers
- dictionary service
- database

### A5. Track progress and provide results

**Inputs:**
- training outcomes
- user activity data
- course/deck completion events

**Controls:**
- progress calculation rules
- completion constraints
- access progression logic

**Outputs:**
- course progress
- deck progress
- card progress
- learning analytics for user

**Mechanisms:**
- progress handlers
- progress services
- database
- frontend progress views

## Recommended Ramus Structure

Create the model in this order:

1. Context diagram `A-0`
2. Child diagram `A0`
3. Blocks `A1` to `A5`

Suggested block names:

- `A1 Authorize user and provide access`
- `A2 Manage learning content`
- `A3 Conduct learning and training`
- `A4 Maintain personal vocabulary`
- `A5 Track progress and provide results`

## Notes

- If required by your teacher, the top-level function can also be named `Organize English learning in the system`.
- If you want, the next step can be a more formal academic version with Russian labels for every ICOM arrow.
