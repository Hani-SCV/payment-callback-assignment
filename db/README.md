# Test database

`docker compose up -d db`를 실행하면 MySQL 8.4가 시작되고 `init` 아래의 SQL이
최초 볼륨 생성 시 한 번 적용됩니다.

시드를 처음 상태로 되돌리려면 다음 명령을 사용합니다.

```bash
docker compose down -v
docker compose up -d db
```

초기화는 해당 Compose 프로젝트의 로컬 볼륨을 삭제합니다. 필요한 개인 데이터가
같은 볼륨에 들어 있지 않은지 확인한 뒤 실행하세요.
