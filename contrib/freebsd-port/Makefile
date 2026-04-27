PORTNAME=	uqda
DISTVERSION=	0.5.13
CATEGORIES=	net

MAINTAINER=	community@uqda.dev
COMMENT=	Fully encrypted self-arranging IPv6 mesh network

WWW=		https://github.com/Uqda/Core

# RC script (files/uqda.in) uses /usr/local paths; patch for PREFIX != /usr/local.

LICENSE=	LGPL3

USES=		go:modules zip
USE_GITHUB=	yes
GH_ACCOUNT=	Uqda
GH_PROJECT=	Core

GO_MODULE=	github.com/Uqda/Core

USE_RC_SUBR=	uqda

GO_BUILDFLAGS=	-trimpath -ldflags="-s -w \
		-X github.com/Uqda/Core/src/version.BuildName=uqda \
		-X github.com/Uqda/Core/src/version.BuildVersion=${DISTVERSION}"

do-build:
	cd ${WRKSRC} && ${SETENV} ${MAKE_ENV} ${GO_ENV} \
		${GO_CMD} build ${GO_BUILDFLAGS} -o ${WRKSRC}/uqda ./cmd/uqda
	cd ${WRKSRC} && ${SETENV} ${MAKE_ENV} ${GO_ENV} \
		${GO_CMD} build ${GO_BUILDFLAGS} -o ${WRKSRC}/uqdactl ./cmd/uqdactl

do-install:
	${MKDIR} ${STAGEDIR}${PREFIX}/sbin
	${INSTALL_PROGRAM} ${WRKSRC}/uqda ${STAGEDIR}${PREFIX}/sbin/uqda
	${INSTALL_PROGRAM} ${WRKSRC}/uqdactl ${STAGEDIR}${PREFIX}/sbin/uqdactl

.include <bsd.port.mk>
