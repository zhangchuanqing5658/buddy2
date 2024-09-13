#include <stdio.h>
#include "buddy2.h"

int main() {
  char cmd[80];
  int arg;
  union buddy_info_union_t info;
  union buddy_info_union_t info2;
  struct buddy2* buddy = buddy2_new(64);
  if (buddy == NULL) {
    printf("malloc failed\n");
    return -1;
  }
  buddy2_dump(buddy);

  info.value  = buddy2_alloc(buddy, 3);
  printf("malloc off:%d, size:%d\n", info._info.offset, info._info.level);
  buddy2_dump(buddy);

  info2.value = buddy2_alloc(buddy, 1);
  printf("malloc off:%d, size:%d\n", info2._info.offset, info2._info.level);
  buddy2_dump(buddy);

  buddy2_free(buddy, info._info.offset);
  printf("free off:%d, size:%d\n", info._info.offset, info._info.level);
  buddy2_dump(buddy);

  buddy2_free(buddy, info2._info.offset);
  printf("free off:%d, size:%d\n", info2._info.offset, info2._info.level);
  buddy2_dump(buddy);
}
