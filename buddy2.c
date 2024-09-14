#include <stdlib.h>
#include <assert.h>
#include <stdio.h>
#include "buddy2.h"

struct buddy2* buddy2_new(int size, int alloc_max_size) {
  struct buddy2* self;
  uint8_t level;

  if (size < 1 || !IS_POWER_OF_2(size) || alloc_max_size < 1 || size < alloc_max_size)
    return NULL;

  self = (struct buddy2*)malloc(sizeof(struct buddy2) + (2 * size - 1) * sizeof(uint8_t));
  self->level = u32log2(size);
  self->node_size = size;
  self->max_level = u32log2(fixsize(alloc_max_size));              //256  page 4K, size 1M
  level = self->level;

  for (int i = 0; i < 2 * size - 1; ++i) {
    if (IS_POWER_OF_2(i+1) && i != 0)
      level--;
    self->bits[i] = level;
  }
  return self;
}

void buddy2_destroy( struct buddy2* self) {
  free(self);
}

uint32_t buddy2_alloc(struct buddy2* self, int size) {
  int32_t index = 0;
  union buddy_info_union_t info;
  int8_t  level = 0;

  size = fixsize(size);
  level = u32log2(size);

  if (self==NULL || self->max_level < level)
    return INVALID_BUDDY_FLAG;

  if (self->bits[index] < level)
    return INVALID_BUDDY_FLAG;

  for(int8_t cur_level = self->level; cur_level != level; cur_level-- ) {
    if (self->bits[LEFT_LEAF(index)] >= level)
      index = LEFT_LEAF(index);
    else
      index = RIGHT_LEAF(index);
  }

  info._info.level  = level;
  info._info.offset = (index + 1) * (1 << level) - self->node_size;             //index1=size/level - 1   index2=off/level   index=index1+index2
  self->bits[index]  = -1;

  while (index) {
    index = PARENT(index);
    self->bits[index] = MAX(self->bits[LEFT_LEAF(index)], self->bits[RIGHT_LEAF(index)]);
  }

  return info.value;
}

void buddy2_free(struct buddy2* self, int offset) {
  unsigned index = 0;
  unsigned left_level, right_level;
  int8_t level = 0;

  assert(self && offset >= 0 && offset < self->node_size);

  level = 0;
  index = offset + self->node_size - 1;       //cal leaf node   index1=size -1 index2=off  index=index1+index2
  self->bits[index] = level;

  while (index) {
    index = PARENT(index);
    left_level  = self->bits[LEFT_LEAF(index)];
    right_level = self->bits[RIGHT_LEAF(index)];

    if (left_level == level && right_level == level)
      self->bits[index] = left_level + 1;
    else
      self->bits[index] = MAX(left_level, right_level);

    level++;
  }
}

void buddy2_dump(struct buddy2* self) {
  int level = 0;
  int count = 0;
  int balck_count = 1 << (self->level - level - 1);

  for (int i = 0; i < 2 * self->node_size - 1; ++i) {
    if (IS_POWER_OF_2(i+1) && i != 0) {
      printf("\t\t\tlevel:%d\t count:%d\t black_count:%d\n", level, count, balck_count);
      level++;
      count = 0;
      balck_count = (1 << (self->level - level - 1));
    }
    count++;
    for (int j = 0; j < balck_count; j++)
      printf("  ");
    if (self->bits[i] >= 0)
      printf("%2d", self->bits[i]);
    else
      printf("%d", self->bits[i]);
  }
  printf("\t\t\tlevel:%d\t count:%d\t\t\n", level, count);
}

// int buddy2_size(struct buddy2* self, int offset) {
//   unsigned node_size, index = 0;

//   assert(self && offset >= 0 && offset < self->node_size);

//   index = offset + self->node_size - 1;
//   //是否需要
//   node_size = 1 << (self->max_level - u32log2(index));

//   // node_size = 1;
//   // for (index = offset + self->size - 1; self->bits[index] ; index = PARENT(index))
//   //   node_size *= 2;

//   return node_size;
// }

// void buddy2_dump(struct buddy2* self) {
//   char canvas[65];
//   int i,j;
//   unsigned node_size, offset;

//   if (self == NULL) {
//     printf("buddy2_dump: (struct buddy2*)self == NULL");
//     return;
//   }

//   if (self->node_size > 64) {
//     printf("buddy2_dump: (struct buddy2*)self is too big to dump");
//     return;
//   }

//   memset(canvas,'_', sizeof(canvas));
//   node_size = self->node_size * 2;

//   for (i = 0; i < 2 * self->node_size - 1; ++i) {
//     if ( IS_POWER_OF_2(i+1) )
//       node_size /= 2;

//     if ( self->bits[i] == 0 ) {
//       if (i >=  self->node_size - 1) {
//         canvas[i - self->node_size + 1] = '*';
//       }
//       else if (self->bits[LEFT_LEAF(i)] && self->bits[RIGHT_LEAF(i)]) {
//         offset = (i+1) * node_size - self->node_size;

//         for (j = offset; j < offset + node_size; ++j)
//           canvas[j] = '*';
//       }
//     }
//   }
//   canvas[self->node_size] = '\0';
//   puts(canvas);
// }