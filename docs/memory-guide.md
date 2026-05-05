# Memory System

MSA's memory system automatically learns from your conversations to provide personalized responses.

## Features

- **Automatic Recording** - All conversations are automatically saved
- **AI Knowledge Extraction** - Extracts preferences, concepts, strategies from your chats
- **Smart Memory Injection** - AI uses your memory to provide personalized responses
- **Memory Browser** - View history, search through conversations and knowledge

## Memory Browser

Open the memory browser from chat:

```
/remember
```

### Browser Sections

1. **History Sessions** - Browse all past conversations
2. **Knowledge Base** - View extracted knowledge (user profile, watchlist, concepts, strategies, Q&A)
3. **Search** - Full-text search across all sessions and knowledge
4. **Statistics** - View usage statistics

## Resuming Sessions

When you exit a chat session, MSA displays a session ID you can use to resume:

```
────────────────────────────────────────
会话已保存: abc-123-def-456
提示: 使用 "msa --resume abc-123-def-456" 恢复此会话
────────────────────────────────────────
```

To resume:

```bash
# Using the session ID from exit message
msa --resume abc-123-def-456

# Or using the short form
msa -r abc-123-def-456
```

When you resume a session:
- Historical messages are loaded into the conversation context
- AI remembers your previous preferences and knowledge
- You can continue the conversation seamlessly
- All extracted knowledge from the session is available

## Privacy & Security

- All data stored locally in `~/.msa/remember/`
- Automatic filtering of sensitive information (API keys, passwords)
- No cloud synchronization - complete privacy

## Disabling Memory

```bash
export MSA_MEMORY_ENABLED=false
msa chat
```
