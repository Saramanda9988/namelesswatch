import type { ImportedGame } from '@/types/game'

export const mockGames: ImportedGame[] = [
  {
    id: 'mock-moonlit-archive',
    title: '月影档案：失落车站',
    importedAt: '2026-05-30T11:10:00.000Z',
    files: {
      'scene.md': '你是一个极度依赖电子手表的高中生。周一傍晚，父母出差，你独自在家。你刚打开玄关灯，手表突然响起一个从未设置过的闹钟。',
      'rule.md': '# 月影档案：失落车站\n\n- 只能根据手表、屋内异常和邻里线索推进。\n- 不要忽略连续出现的闹钟。\n- 如果用户反复无视安全提示，故事应走向循环或坏结局。\n- 每轮给出 2 到 4 个可执行选择。',
      'true.md': '手表记录的是上一轮循环留下的提醒。真正危险来自厨房冰箱后的破损电源线和被遗忘的煤气阀。',
      'memory.md': '阶段：周一傍晚刚到家。\n状态：父母出差三天，你独自在家；手表响起未知闹钟；尚未做出行动。\n行动记录：无。',
      'endings.md': '- good_safe: 安全结局，用户及时处理异常并求助。\n- bad_fire: 坏结局，用户忽略厨房危险。\n- loop_watch: 循环结局，用户反复无视手表提醒。',
    },
    photoUrls: [
      'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=900&q=80',
      'https://images.unsplash.com/photo-1473448912268-2022ce9509d8?auto=format&fit=crop&w=1200&q=80',
    ],
    mapUrls: [
      'https://images.unsplash.com/photo-1524661135-423995f22d0b?auto=format&fit=crop&w=700&q=80',
      'https://images.unsplash.com/photo-1531259683007-016a7b628fc3?auto=format&fit=crop&w=700&q=80',
    ],
    script: [
      {
        id: 'line-001',
        speaker: '旁白',
        text: '凌晨两点十七分，旧车站的候车厅仍亮着一盏灯。雨水沿着玻璃顶棚滑落，像有人在黑暗里反复擦拭一段被遗忘的记忆。',
        backgroundUrl:
          'https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=1800&q=80',
      },
      {
        id: 'line-002',
        speaker: '林澈',
        text: '这张车票没有目的地，只有一个编号。可我总觉得，只要走上站台，就会有人在那里等我。',
        backgroundUrl:
          'https://images.unsplash.com/photo-1519608487953-e999c86e7455?auto=format&fit=crop&w=1800&q=80',
      },
      {
        id: 'line-003',
        speaker: '陌生少女',
        text: '你迟到了。上一轮也是这样。别再问我是谁，先把钟楼的门打开。',
        backgroundUrl:
          'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?auto=format&fit=crop&w=1800&q=80',
      },
      {
        id: 'line-004',
        speaker: '旁白',
        text: '她把一枚铜钥匙放进你的掌心。钥匙冰冷，却在接触皮肤的瞬间浮现出细小的蓝光。',
        backgroundUrl:
          'https://images.unsplash.com/photo-1448375240586-882707db888b?auto=format&fit=crop&w=1800&q=80',
      },
      {
        id: 'line-005',
        speaker: '林澈',
        text: '如果我打开那扇门，一切都会结束吗？',
        backgroundUrl:
          'https://images.unsplash.com/photo-1500534623283-312aade485b7?auto=format&fit=crop&w=1800&q=80',
      },
      {
        id: 'line-006',
        speaker: '陌生少女',
        text: '不。只是终于开始。',
        backgroundUrl:
          'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?auto=format&fit=crop&w=1800&q=80',
      },
    ],
  },
]
